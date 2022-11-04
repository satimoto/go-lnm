package channelevent

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-lsp/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelEventMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	HtlcMonitor            *htlc.HtlcMonitor
	ChannelEventsClient    lnrpc.Lightning_SubscribeChannelEventsClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	NodeRepository         node.NodeRepository
	UserResolver           *user.UserResolver
	nodeID                 int64
}

func NewChannelEventMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver, htlcMonitor *htlc.HtlcMonitor) *ChannelEventMonitor {
	return &ChannelEventMonitor{
		LightningService:       services.LightningService,
		HtlcMonitor:            htlcMonitor,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeRepository:         node.NewRepository(repositoryService),
		UserResolver:           user.NewResolver(repositoryService, services),
	}
}

func (m *ChannelEventMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Channel Events")
	channelEventChan := make(chan lnrpc.ChannelEventUpdate)

	m.nodeID = nodeID
	go m.waitForChannelEvents(shutdownCtx, waitGroup, channelEventChan)
	go m.subscribeChannelEvents(channelEventChan)
}

func (m *ChannelEventMonitor) handleChannelEvent(channelEvent lnrpc.ChannelEventUpdate) {
	/** Channel Event received.
	 *  Find the Channel Request by the channel point params.
	 *  Update the Channel Request status depending on the event type.
	 */
	log.Printf("Channel Event: %v", channelEvent.Type)

	ctx := context.Background()

	// TODO: restrict user if all channels are closed
	switch channelEvent.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		/** Channel Open.
		 *  Set the channel request Txid and OutputIndex.
		 *  Unrestrict the user's token to allow charging.
		 */
		openChannel := channelEvent.GetOpenChannel()
		txidBytes, outputIndex, _ := util.ConvertChannelPoint(openChannel.ChannelPoint)
		txid := hex.EncodeToString(util.ReverseBytes(txidBytes))
		log.Printf("Txid: %v", txid)
		log.Printf("OutputIndex: %v", outputIndex)

		updatePendingChannelRequestByPubkeyParams := db.UpdatePendingChannelRequestByPubkeyParams{
			Pubkey:           openChannel.RemotePubkey,
			FundingAmount:    dbUtil.SqlNullInt64(openChannel.Capacity),
			FundingTxID:      dbUtil.SqlNullString(txid),
			FundingTxIDBytes: txidBytes,
			OutputIndex:      dbUtil.SqlNullInt64(outputIndex),
		}

		if channelRequest, err := m.ChannelRequestResolver.Repository.UpdatePendingChannelRequestByPubkey(ctx, updatePendingChannelRequestByPubkeyParams); err == nil {
			user, err := m.UserResolver.Repository.GetUser(ctx, channelRequest.UserID)

			if err != nil {
				metrics.RecordError("LSP048", "Error retieving channel request user", err)
				log.Printf("LSP048: ChannelRequestID=%v, UserID=%v", channelRequest.ID, channelRequest.UserID)
				return
			}

			err = m.UserResolver.UnrestrictUser(ctx, user)

			if err != nil {
				metrics.RecordError("LSP049", "Error unrestricting user", err)
				log.Printf("LSP049: ChannelRequestID=%v, UserID=%v", channelRequest.ID, channelRequest.UserID)
			}
		}

		m.updateNode(ctx)
	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		m.updateNode(ctx)
	}
}

func (m *ChannelEventMonitor) subscribeChannelEvents(channelEventChan chan<- lnrpc.ChannelEventUpdate) {
	channelEventsClient, err := m.waitForSubscribeChannelEventsClient(0, 1000)
	dbUtil.PanicOnError("LSP012", "Error creating Channel Events client", err)

	m.ChannelEventsClient = channelEventsClient

	for {
		channelEvent, err := m.ChannelEventsClient.Recv()

		if err == nil {
			channelEventChan <- *channelEvent
		} else {
			m.ChannelEventsClient, err = m.waitForSubscribeChannelEventsClient(100, 1000)
			dbUtil.PanicOnError("LSP013", "Error creating Channel Events client", err)
		}
	}
}

func (m *ChannelEventMonitor) updateNode(ctx context.Context) {
	getInfoResponse, err := m.LightningService.GetInfo(&lnrpc.GetInfoRequest{})

	if err != nil {
		metrics.RecordError("LSP077", "Error getting info", err)
	}

	n, err := m.NodeRepository.GetNode(ctx, m.nodeID)

	if err != nil {
		metrics.RecordError("LSP078", "Error getting info", err)
		log.Printf("LSP078: NodeID=%v", m.nodeID)
	}

	numChannels := int64(getInfoResponse.NumActiveChannels + getInfoResponse.NumInactiveChannels + getInfoResponse.NumPendingChannels)

	updateNodeParams := param.NewUpdateNodeParams(n)
	updateNodeParams.Channels = numChannels
	updateNodeParams.Peers = int64(getInfoResponse.NumPeers)

	_, err = m.NodeRepository.UpdateNode(ctx, updateNodeParams)

	if err != nil {
		metrics.RecordError("LSP079", "Error updating node", err)
		log.Printf("LSP079: Params=%#v", updateNodeParams)
	}

	metricChannelsTotal.Set(float64(numChannels))
}

func (m *ChannelEventMonitor) waitForChannelEvents(shutdownCtx context.Context, waitGroup *sync.WaitGroup, channelEventChan chan lnrpc.ChannelEventUpdate) {
	waitGroup.Add(1)
	defer close(channelEventChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Channel Events")
			return
		case channelEvent := <-channelEventChan:
			m.handleChannelEvent(channelEvent)
		}
	}
}

func (m *ChannelEventMonitor) waitForSubscribeChannelEventsClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeChannelEventsClient, err := m.LightningService.SubscribeChannelEvents(&lnrpc.ChannelEventSubscription{})

		if err == nil {
			return subscribeChannelEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Channel Events client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
