package commands

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	cfg "github.com/adriamb/gotoma/config"
	"github.com/adriamb/gotoma/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)

type EventLogFunc func(netid, alert, message string, tx *types.Transaction, r *types.Receipt)

type EthNetwork struct {
	id      string
	storage *KVStorage
	client  *ethclient.Client
	scanner *eth.ScanEventDispatcher
	logfunc EventLogFunc
}

func NewEthNetwork(netid, url string, storage *KVStorage, logfunc EventLogFunc) *EthNetwork {
	client, err := ethclient.Dial(url)
	assert(err)

	ethnetwork := &EthNetwork{netid, storage, client, nil, logfunc}

	ethnetwork.scanner = eth.NewScanEventDispatcher(
		client, ethnetwork.Process, ethnetwork, netid,
	)
	return ethnetwork
}

func (s *EthNetwork) Start() error {
	s.scanner.Start()
	return nil
}

func (s *EthNetwork) Stop() {
	s.scanner.Stop()
	s.scanner.Join()
}

func (s *EthNetwork) Load() (lastBlock uint64, lastTxIndex uint, err error) {

	// return 6293857, 112, nil

	lastBlockStr, ok := s.storage.Get(s.id + ".lastBlock")
	if !ok {
		block, err := s.client.BlockByNumber(context.TODO(), nil)
		if err != nil {
			return 0, 0, err
		}
		return block.NumberU64(), 0, nil
	}

	lastTxIndexStr, ok := s.storage.Get(s.id + ".lastTxIndex")
	if !ok {
		return 0, 0, errors.New("Bad configuration")
	}

	lastBlockInt, err := strconv.Atoi(lastBlockStr)
	if err != nil {
		return 0, 0, err
	}
	lastTxIndexInt, err := strconv.Atoi(lastTxIndexStr)
	if err != nil {
		return 0, 0, err
	}

	return uint64(lastBlockInt), uint(lastTxIndexInt), nil
}

func (s *EthNetwork) Save(lastBlock uint64, lastTxIndex uint) error {
	if err := s.storage.Put(s.id+".lastBlock", strconv.Itoa(int(lastBlock))); err != nil {
		return err
	}
	return s.storage.Put(s.id+".lastTxIndex", strconv.Itoa(int(lastTxIndex)))
}

func (s *EthNetwork) Process(tx *types.Transaction, r *types.Receipt) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := signer.Sender(tx)
	if err != nil {
		return err
	}

	params := make(map[string]interface{}, 8)
	params["from"] = strings.ToLower(from.Hex())
	if tx.To() != nil {
		params["to"] = strings.ToLower(tx.To().Hex())
	} else {
		params["to"] = "create"
	}
	params["value"] = tx.Value()
	params["gas"] = tx.Gas()
	params["gasprice"] = tx.GasPrice()
	params["data"] = "0x" + hex.EncodeToString(tx.Data())

	for _, l := range r.Logs {
		k := "log_" + l.Address.Hex() + "_" + l.Topics[0].Hex()
		params[k] = "0x" + hex.EncodeToString(l.Data)
	}

	for acc, accdata := range cfg.C.Accounts {
		if accdata.Network != s.id {
			continue
		}

		acc = strings.ToLower(strings.TrimSpace(acc))

		if acc != params["from"] && acc != params["to"] {
			continue
		}

		s.logfunc(s.id, "Generic", acc+" account modified", tx, r)
		continue
	}

	for alertid, alert := range cfg.C.Alerts {
		if alert.Network != s.id {
			continue
		}
		expr, err := govaluate.NewEvaluableExpression(alert.Rule)
		if err != nil {
			return err
		}
		result, err := expr.Evaluate(params)
		if err != nil {
			if strings.HasPrefix(err.Error(), "No parameter 'log") {
				continue
			}
			return err
		}
		boolresult, ok := result.(bool)
		if !ok {
			return fmt.Errorf("Should return bool")
		}
		if boolresult {
			tmpl, err := template.New("").Parse(alert.Message)
			if err != nil {
				return err
			}
			var message bytes.Buffer
			if err = tmpl.Execute(&message, params); err != nil {
				return err
			}
			s.logfunc(s.id, alertid, string(message.Bytes()), tx, r)
		}
	}

	return nil
}

func (s *EthNetwork) TxInfo(txid string) string {
	tx, _, err := s.client.TransactionByHash(context.TODO(), common.HexToHash(txid))
	if err != nil {
		return fmt.Sprintf("error retriving tx: %v", err)
	}
	receipt, err := s.client.TransactionReceipt(context.TODO(), common.HexToHash(txid))
	if err != nil {
		return fmt.Sprintf("error retriving receipt: %v", err)
	}

	return htmlTx(tx, receipt)
}

type Network interface {
	Start() error
	TxInfo(tx string) string
	Stop()
}
