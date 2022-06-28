package blockepoch

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BlockEpochMonitor struct {
	LightningService  lightningnetwork.LightningNetwork
	BlockEpochsClient chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient
	nodeID            int64
}

func NewBlockEpochMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *BlockEpochMonitor {
	return &BlockEpochMonitor{
		LightningService: lightningService,
	}
}

func (m *BlockEpochMonitor) StartMonitor(nodeID int64, ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Block Epochs")
	blockEpochChan := make(chan chainrpc.BlockEpoch)

	m.nodeID = nodeID
	go m.waitForBlockEpochs(ctx, waitGroup, blockEpochChan)
	go m.subscribeBlockEpochNotifications(blockEpochChan)
}

func (m *BlockEpochMonitor) handleBlockEpoch(blockEpoch chainrpc.BlockEpoch) {
	/** Block Epoch received.
	 *
	 */

	log.Printf("Hash: %v", hex.EncodeToString(blockEpoch.Hash))
	log.Printf("Height: %v", blockEpoch.Height)
}

func (m *BlockEpochMonitor) subscribeBlockEpochNotifications(blockEpochChan chan<- chainrpc.BlockEpoch) {
	htlcEventsClient, err := m.waitForRegisterBlockEpochNtfnClient(0, 1000)
	util.PanicOnError("LSP073", "Error creating Block Epochs client", err)
	m.BlockEpochsClient = htlcEventsClient

	for {
		blockEpoch, err := m.BlockEpochsClient.Recv()

		if err == nil {
			blockEpochChan <- *blockEpoch
		} else {
			m.BlockEpochsClient, err = m.waitForRegisterBlockEpochNtfnClient(100, 1000)
			util.PanicOnError("LSP074", "Error creating Block Epochs client", err)
		}
	}
}

func (m *BlockEpochMonitor) waitForBlockEpochs(ctx context.Context, waitGroup *sync.WaitGroup, blockEpochChan chan chainrpc.BlockEpoch) {
	waitGroup.Add(1)
	defer close(blockEpochChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Block Epochs")
			return
		case blockEpoch := <-blockEpochChan:
			m.handleBlockEpoch(blockEpoch)
		}
	}
}

func (m *BlockEpochMonitor) waitForRegisterBlockEpochNtfnClient(initialDelay, retryDelay time.Duration) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		registerBlockEpochNtfnClient, err := m.LightningService.RegisterBlockEpochNtfn(&chainrpc.BlockEpoch{})

		if err == nil {
			return registerBlockEpochNtfnClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Block Epochs client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
