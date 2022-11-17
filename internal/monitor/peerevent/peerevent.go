package peerevent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/internal/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PeerEventMonitor struct {
	LightningService lightningnetwork.LightningNetwork
	PeerEventsClient lnrpc.Lightning_SubscribePeerEventsClient
	UserResolver     *user.UserResolver
	nodeID           int64
}

func NewPeerEventMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *PeerEventMonitor {
	return &PeerEventMonitor{
		LightningService: services.LightningService,
		UserResolver:     user.NewResolver(repositoryService, services),
	}
}

func (m *PeerEventMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Peer Events")
	peerEventChan := make(chan lnrpc.PeerEvent)

	m.nodeID = nodeID
	go m.waitForPeerEvents(shutdownCtx, waitGroup, peerEventChan)
	go m.subscribePeerEventInterceptions(peerEventChan)
}

func (m *PeerEventMonitor) handlePeerEvent(peerEvent lnrpc.PeerEvent) {
	log.Printf("Pubkey: %v", peerEvent.PubKey)
	log.Printf("Type: %v", peerEvent.Type)

	/** PeerEvent received.
	 *  Update the user last active date
	 */

	ctx := context.Background()

	updateUserByPubkeyParams := db.UpdateUserByPubkeyParams{
		Pubkey:         peerEvent.PubKey,
		LastActiveDate: util.SqlNullTime(time.Now()),
	}

	m.UserResolver.Repository.UpdateUserByPubkey(ctx, updateUserByPubkeyParams)
}

func (m *PeerEventMonitor) subscribePeerEventInterceptions(peerEventChan chan<- lnrpc.PeerEvent) {
	peerEventsClient, err := m.waitForSubscribePeerEventsClient(0, 1000)
	util.PanicOnError("LSP129", "Error creating Peer Events client", err)
	m.PeerEventsClient = peerEventsClient

	for {
		peerEvent, err := m.PeerEventsClient.Recv()

		if err == nil {
			peerEventChan <- *peerEvent
		} else {
			m.PeerEventsClient, err = m.waitForSubscribePeerEventsClient(100, 1000)
			util.PanicOnError("LSP129", "Error creating PeerEvents client", err)
		}
	}
}

func (m *PeerEventMonitor) waitForPeerEvents(shutdownCtx context.Context, waitGroup *sync.WaitGroup, peerEventChan chan lnrpc.PeerEvent) {
	waitGroup.Add(1)
	defer close(peerEventChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Peer Events")
			return
		case peerEvent := <-peerEventChan:
			m.handlePeerEvent(peerEvent)
		}
	}
}

func (m *PeerEventMonitor) waitForSubscribePeerEventsClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribePeerEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribePeerEventsClient, err := m.LightningService.SubscribePeerEvents(&lnrpc.PeerEventSubscription{})

		if err == nil {
			return subscribePeerEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Peer Events client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
