package channelgraph

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelGraphMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	HtlcMonitor            *htlc.HtlcMonitor
	ChannelGraphClient     lnrpc.Lightning_SubscribeChannelGraphClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	NodeRepository         node.NodeRepository
	UserResolver           *user.UserResolver
	nodeID                 int64
}

func NewChannelGraphMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, htlcMonitor *htlc.HtlcMonitor) *ChannelGraphMonitor {
	return &ChannelGraphMonitor{
		LightningService:       lightningService,
		HtlcMonitor:            htlcMonitor,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeRepository:         node.NewRepository(repositoryService),
		UserResolver:           user.NewResolver(repositoryService),
	}
}

func (m *ChannelGraphMonitor) StartMonitor(nodeID int64, ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Channel Graph")
	channelGraphChan := make(chan lnrpc.GraphTopologyUpdate)

	m.nodeID = nodeID
	go m.waitForChannelGraphs(ctx, waitGroup, channelGraphChan)
	go m.subscribeChannelGraphs(channelGraphChan)
}

func (m *ChannelGraphMonitor) handleChannelGraph(channelGraph lnrpc.GraphTopologyUpdate) {
	/** Channel Graph received.
	 *  Find the Channel Request by the channel point params.
	 *  Update the Channel Request status depending on the event type.
	 */
	log.Printf("Channel Graph: %v", len(channelGraph.ChannelUpdates))
	ctx := context.Background()

	for _, channelUpdate := range channelGraph.ChannelUpdates {
		txidBytes := channelUpdate.ChanPoint.GetFundingTxidBytes()
		outputIndex := channelUpdate.ChanPoint.OutputIndex

		getChannelRequestByChannelPointParams := db.GetChannelRequestByChannelPointParams{
			FundingTxID: txidBytes,
			OutputIndex: dbUtil.SqlNullInt64(outputIndex),
		}

		channelRequest, err := m.ChannelRequestResolver.Repository.GetChannelRequestByChannelPoint(ctx, getChannelRequestByChannelPointParams)

		if err == nil && channelRequest.Status == db.ChannelRequestStatusOPENINGCHANNEL {
			shortChanID := lnwire.NewShortChanIDFromInt(channelUpdate.ChanId)
			txid := hex.EncodeToString(util.ReverseBytes(txidBytes))
			
			log.Printf("ChanID: %v", shortChanID.String())
			log.Printf("Txid: %v", txid)
			log.Printf("OutputIndex: %v", outputIndex)
			log.Printf("AdvertisingNode: %v", channelUpdate.AdvertisingNode)
			log.Printf("ConnectingNode: %v", channelUpdate.ConnectingNode)
			log.Printf("Disabled: %v", channelUpdate.RoutingPolicy.Disabled)	

			go m.HtlcMonitor.ResumeChannelRequestHtlcs(channelRequest)
		}
	}
}

func (m *ChannelGraphMonitor) subscribeChannelGraphs(channelGraphChan chan<- lnrpc.GraphTopologyUpdate) {
	channelGraphsClient, err := m.waitForSubscribeChannelGraphClient(0, 1000)
	dbUtil.PanicOnError("LSP104", "Error creating Channel Graph client", err)

	m.ChannelGraphClient = channelGraphsClient

	for {
		channelGraph, err := m.ChannelGraphClient.Recv()

		if err == nil {
			channelGraphChan <- *channelGraph
		} else {
			m.ChannelGraphClient, err = m.waitForSubscribeChannelGraphClient(100, 1000)
			dbUtil.PanicOnError("LSP105", "Error creating Channel Graph client", err)
		}
	}
}

func (m *ChannelGraphMonitor) waitForChannelGraphs(ctx context.Context, waitGroup *sync.WaitGroup, channelGraphChan chan lnrpc.GraphTopologyUpdate) {
	waitGroup.Add(1)
	defer close(channelGraphChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Channel Graph")
			return
		case channelGraph := <-channelGraphChan:
			m.handleChannelGraph(channelGraph)
		}
	}
}

func (m *ChannelGraphMonitor) waitForSubscribeChannelGraphClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeChannelGraphClient, err := m.LightningService.SubscribeChannelGraph(&lnrpc.GraphTopologySubscription{})

		if err == nil {
			return subscribeChannelGraphClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Channel Graph client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
