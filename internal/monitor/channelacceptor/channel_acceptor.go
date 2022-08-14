package channelacceptor

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	lnwire "github.com/lightningnetwork/lnd/channeldb/migration/lnwire21"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelAcceptorMonitor struct {
	LightningService      lightningnetwork.LightningNetwork
	ChannelAcceptorClient lnrpc.Lightning_ChannelAcceptorClient
	nodeID                int64
}

func NewChannelAcceptorMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *ChannelAcceptorMonitor {
	return &ChannelAcceptorMonitor{
		LightningService: lightningService,
	}
}

func (m *ChannelAcceptorMonitor) StartMonitor(nodeID int64, ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Channel Acceptor")
	channelAcceptorChan := make(chan lnrpc.ChannelAcceptRequest)

	m.nodeID = nodeID
	go m.waitForChannelAcceptor(ctx, waitGroup, channelAcceptorChan)
	go m.subscribeChannelAcceptor(channelAcceptorChan)
}

func (m *ChannelAcceptorMonitor) handleChannelAcceptor(channelAccept lnrpc.ChannelAcceptRequest) {
	/** Channel Accept received.
	 *  Reject a channel request that is zero-conf or private
	 */
	wantsPrivate := channelAccept.ChannelFlags&uint32(lnwire.FFAnnounceChannel) == 0
	wantsZeroConf := false
	wantsScidAlias := false
	
	log.Print("Channel Accept")
	log.Printf("NodePubkey: %v", hex.EncodeToString(channelAccept.NodePubkey))
	log.Printf("WantsPrivate: %v", wantsPrivate)
	log.Printf("WantsZeroConf: %v", wantsZeroConf)
	log.Printf("WantsScidAlias: %v", wantsScidAlias)

	m.ChannelAcceptorClient.Send(&lnrpc.ChannelAcceptResponse{
		PendingChanId: channelAccept.PendingChanId,
		Accept:        !wantsPrivate && !wantsZeroConf,
	})
}

func (m *ChannelAcceptorMonitor) subscribeChannelAcceptor(channelAcceptorChan chan<- lnrpc.ChannelAcceptRequest) {
	channelAcceptorsClient, err := m.waitForChannelAcceptorClient(0, 1000)
	dbUtil.PanicOnError("LSP012", "Error creating Channel Acceptor client", err)

	m.ChannelAcceptorClient = channelAcceptorsClient

	for {
		channelAcceptor, err := m.ChannelAcceptorClient.Recv()

		if err == nil {
			channelAcceptorChan <- *channelAcceptor
		} else {
			m.ChannelAcceptorClient, err = m.waitForChannelAcceptorClient(100, 1000)
			dbUtil.PanicOnError("LSP013", "Error creating Channel Acceptor client", err)
		}
	}
}

func (m *ChannelAcceptorMonitor) waitForChannelAcceptor(ctx context.Context, waitGroup *sync.WaitGroup, channelAcceptorChan chan lnrpc.ChannelAcceptRequest) {
	waitGroup.Add(1)
	defer close(channelAcceptorChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Channel Acceptor")
			return
		case channelAcceptor := <-channelAcceptorChan:
			m.handleChannelAcceptor(channelAcceptor)
		}
	}
}

func (m *ChannelAcceptorMonitor) waitForChannelAcceptorClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_ChannelAcceptorClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeChannelAcceptorClient, err := m.LightningService.ChannelAcceptor()

		if err == nil {
			return subscribeChannelAcceptorClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Channel Acceptor client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
