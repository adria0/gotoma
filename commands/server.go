package commands

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	cfg "github.com/TomahawkEthBerlin/gotoma/config"
	"github.com/TomahawkEthBerlin/gotoma/eth"
	"github.com/ethereum/go-ethereum/core/types"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

const (
	kEthereum = "ethereum"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

type EthNetwork struct {
	id      string
	storage *KVStorage
	client  *ethclient.Client
	scanner *eth.ScanEventDispatcher
}

func NewEthNetwork(id, url string, storage *KVStorage) *EthNetwork {
	client, err := ethclient.Dial(url)
	assert(err)

	ethnetwork := &EthNetwork{id, storage, client, nil}

	ethnetwork.scanner = eth.NewScanEventDispatcher(
		client, ethnetwork.Process, ethnetwork,
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

	return 6293857, 112, nil

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

func (s *EthNetwork) EmitAlert(alert string, message string, tx *types.Transaction) {

	log.Info("Alert type " + alert + " : " + message + " (tx " + tx.Hash().Hex())
}

func (s *EthNetwork) Process(tx *types.Transaction, r *types.Receipt) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := signer.Sender(tx)
	if err != nil {
		return err
	}

	params := make(map[string]interface{}, 8)
	params["from"] = strings.ToLower(from.Hex())
	params["to"] = strings.ToLower(tx.To().Hex())
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

		s.EmitAlert("Generic", acc+" account modified", tx)
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
			s.EmitAlert(alertid, string(message.Bytes()), tx)
		}
	}

	return nil
}

type KVStorage struct {
	sync.Mutex
	values   map[string]string
	filename string
}

func NewKVStorage(filename string) (*KVStorage, error) {

	values := make(map[string]string)

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || len(trimmed) == 0 {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		values[kv[0]] = kv[1]
	}

	return &KVStorage{
		values:   values,
		filename: filename,
	}, nil
}

func (kv *KVStorage) Get(k string) (string, bool) {
	kv.Lock()
	value, ok := kv.values[k]
	kv.Unlock()
	return value, ok
}
func (kv *KVStorage) Put(k, v string) error {
	kv.Lock()
	defer kv.Unlock()

	kv.values[k] = v

	var buffer bytes.Buffer
	for k, v = range kv.values {
		buffer.WriteString(fmt.Sprintf("%v=%v\n", k, v))
	}

	return ioutil.WriteFile(kv.filename, buffer.Bytes(), 0744)
}

type Network interface {
	Start() error
	Stop()
}

func Serve(cmd *cobra.Command, args []string) {

	networks := make(map[string]Network)

	storage, err := NewKVStorage("./state")
	assert(err)

	for networkid, netinfo := range cfg.C.Networks {
		if netinfo.Type != kEthereum {
			assert(fmt.Errorf("Unknown network type '%v'", netinfo.Type))
		}
		networks[networkid] = NewEthNetwork(networkid, netinfo.URL, storage)

	}

	for networkid, network := range networks {
		log.Info("Starting dispatcher for ", networkid)
		network.Start()
	}

	time.Sleep(time.Second * 1000)
}
