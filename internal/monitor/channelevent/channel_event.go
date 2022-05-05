package channelevent

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	dbUtil "github.com/satimoto/go-datastore/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelEventMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	ChannelEventsClient    lnrpc.Lightning_SubscribeChannelEventsClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	UserResolver           *user.UserResolver
}

func NewChannelEventMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *ChannelEventMonitor {
	return &ChannelEventMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		UserResolver:           user.NewResolver(repositoryService),
	}
}

func (m *ChannelEventMonitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Channel Events")
	channelEventChan := make(chan lnrpc.ChannelEventUpdate)

	go m.waitForChannelEvents(ctx, waitGroup, channelEventChan)
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
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		pendingOpenChannel := channelEvent.GetPendingOpenChannel()
		log.Printf("Txid: %v", hex.EncodeToString(pendingOpenChannel.Txid))
		log.Printf("OutputIndex: %v", pendingOpenChannel.OutputIndex)

		m.ChannelRequestResolver.Repository.UpdateChannelRequestByChannelPoint(ctx, db.UpdateChannelRequestByChannelPointParams{
			FundingTxID: pendingOpenChannel.Txid,
			OutputIndex: dbUtil.SqlNullInt64(pendingOpenChannel.OutputIndex),
			Status:      db.ChannelRequestStatusOPENINGCHANNEL,
		})
		break
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		openChannel := channelEvent.GetOpenChannel()
		txid, outputIndex, _ := util.ConvertChannelPoint(openChannel.ChannelPoint)
		log.Printf("Txid: %v", hex.EncodeToString(txid))
		log.Printf("OutputIndex: %v", outputIndex)
		
		updateChannelRequestByChannelPointParams := db.UpdateChannelRequestByChannelPointParams{
			FundingTxID: txid,
			OutputIndex: dbUtil.SqlNullInt64(outputIndex),
			Status:      db.ChannelRequestStatusCOMPLETED,
		}

		channelRequest, err := m.ChannelRequestResolver.Repository.UpdateChannelRequestByChannelPoint(ctx, updateChannelRequestByChannelPointParams)

		if err != nil {
			dbUtil.LogOnError("LSP047", "Error updating channel request", err)
			log.Printf("LSP047: Params=%#v", updateChannelRequestByChannelPointParams)
			return
		}

		user, err := m.UserResolver.Repository.GetUser(ctx, channelRequest.UserID)

		if err != nil {
			dbUtil.LogOnError("LSP048", "Error retieving channel request user", err)
			log.Printf("LSP048: ChannelRequestID=%v, UserID=%v", channelRequest.ID, channelRequest.UserID)
			return
		}

		err = m.UserResolver.UnrestrictUser(ctx, user)
		
		if err != nil {
			dbUtil.LogOnError("LSP049", "Error unrestricting user", err)
			log.Printf("LSP049: ChannelRequestID=%v, UserID=%v", channelRequest.ID, channelRequest.UserID)
		}
	
		break
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

func (m *ChannelEventMonitor) waitForChannelEvents(ctx context.Context, waitGroup *sync.WaitGroup, channelEventChan chan lnrpc.ChannelEventUpdate) {
	waitGroup.Add(1)
	defer close(channelEventChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
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
