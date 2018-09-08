package eth

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/sirupsen/logrus"
)

type EventHandlerFunc func(*types.Transaction, *types.Receipt) error

type SavePoint interface {
	Load() (lastBlock uint64, lastTxIndex uint, err error)
	Save(lastBlock uint64, lastTxIndex uint) error
}

type ScanEventDispatcher struct {
	sync.Mutex

	client *ethclient.Client

	netid string

	eventHandler EventHandlerFunc
	savepoint    SavePoint

	block    *types.Block
	receipts *ReceiptDownloader

	terminatech  chan interface{}
	terminatedch chan interface{}

	nextBlock   uint64
	nextTxIndex uint
}

func NewScanEventDispatcher(client *ethclient.Client, handler EventHandlerFunc, savepoint SavePoint, netid string) *ScanEventDispatcher {

	return &ScanEventDispatcher{
		netid:        netid,
		client:       client,
		savepoint:    savepoint,
		eventHandler: handler,

		receipts: NewReceiptDownloader(client, 3),

		terminatech:  make(chan interface{}),
		terminatedch: make(chan interface{}),
	}
}

func (e *ScanEventDispatcher) process() (bool, error) {

	var err error

	// Retrieve the last processed log, this is only called in the first loop.
	if e.nextBlock == 0 {
		e.nextBlock, e.nextTxIndex, err = e.savepoint.Load()
		if err != nil {
			return false, err
		}
		e.nextTxIndex++
	}

	log.WithFields(log.Fields{
		"block/tx": fmt.Sprintf("%v/%v", e.nextBlock, e.nextTxIndex),
		"netid":    e.netid,
	}).Debug("EVENT scanning")

	// Check if e.block is valid, if not download it.
	if e.block == nil || e.block.NumberU64() < e.nextBlock {
		e.block, err = e.client.BlockByNumber(context.TODO(), big.NewInt(int64(e.nextBlock)))

		// Check if block is available, if is in the main chain.
		if err == ethereum.NotFound {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		log.WithFields(log.Fields{
			"block": fmt.Sprintf("%v", e.nextBlock),
			"netid": e.netid,
		}).Info("EVENT processing block")

		// Download all receipts, in parallel
		for index := e.nextTxIndex; index < uint(len(e.block.Transactions())); index++ {
			e.receipts.Request(e.block.Transactions()[index].Hash())
		}
	}

	var receipt *types.Receipt

	// Download the receipt that contains the log
	if e.nextTxIndex < uint(len(e.block.Transactions())) {
		tx := e.block.Transactions()[e.nextTxIndex]
		txid := e.block.Transactions()[e.nextTxIndex].Hash()
		receipt, err = e.receipts.Get(txid)
		e.receipts.Forget(txid)
		err = e.eventHandler(tx, receipt)
		if err != nil {
			log.Error("Failed handler ", err)
		}
		e.nextTxIndex++
	}

	if e.nextTxIndex >= uint(len(e.block.Transactions())) {
		e.nextTxIndex = 0
		e.nextBlock++
	}

	e.savepoint.Save(e.nextBlock, e.nextTxIndex)

	return false, nil
}

// Stop scanning the blockchain for events
func (e *ScanEventDispatcher) Stop() {
	go func() {
		e.terminatech <- nil
	}()
}

// Join waits all background jobs finished
func (e *ScanEventDispatcher) Join() {
	<-e.terminatedch
}

// Start scanning the blockchain for events
func (e *ScanEventDispatcher) Start() {

	go func() {
		e.receipts.Start()
		loop := true
		for loop {
			select {

			case <-e.terminatech:
				log.Debug("EVENT Dispatching terminatech")
				loop = false
				break

			default:
				wait, err := e.process()
				if err != nil {
					log.Error("EVENT Failed ", err, " netid=", e.netid)
					loop = false
				} else if wait {
					time.Sleep(4 * time.Second)
				}
			}
		}
		e.terminatedch <- nil
		e.receipts.Stop()
		e.receipts.Join()

	}()
}
