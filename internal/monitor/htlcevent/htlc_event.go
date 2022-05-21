package htlcevent

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HtlcEventMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	HtlcEventsClient       routerrpc.Router_SubscribeHtlcEventsClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
}

func NewHtlcEventMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *HtlcEventMonitor {
	return &HtlcEventMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}

func (m *HtlcEventMonitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Htlc Events")
	htlcEventChan := make(chan routerrpc.HtlcEvent)

	go m.waitForHtlcEvents(ctx, waitGroup, htlcEventChan)
	go m.subscribeHtlcEventInterceptions(htlcEventChan)
}

func (m *HtlcEventMonitor) handleHtlcEvent(htlcEvent routerrpc.HtlcEvent) {
	log.Printf("HTLC Event: %v", htlcEvent.EventType)

	/** HTLC Event received.
	 *  Check that the event type is a Forward event and that is is successful.
	 *  Find the Channel Request HTLC by the circuit key params.
	 *  If the Channel Request HTLC exists, set it as settled.
	 */

	if htlcEvent.EventType == routerrpc.HtlcEvent_FORWARD {
		successEvent := htlcEvent.GetSettleEvent()
		forwardEvent := htlcEvent.GetForwardEvent()

		if forwardEvent != nil {
			log.Printf("Forward HTLC")
			log.Printf("IncomingAmtMsat: %v", forwardEvent.Info.IncomingAmtMsat)
			log.Printf("OutgoingAmtMsat: %v", forwardEvent.Info.OutgoingAmtMsat)
		}

		if successEvent != nil {
			ctx := context.Background()
			getChannelRequestHtlcByCircuitKeyParams := db.GetChannelRequestHtlcByCircuitKeyParams{
				ChanID: int64(htlcEvent.IncomingChannelId),
				HtlcID: int64(htlcEvent.IncomingHtlcId),
			}

			if channelRequestHtlc, err := m.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err == nil {
				m.ChannelRequestResolver.Repository.UpdateChannelRequestHtlc(ctx, db.UpdateChannelRequestHtlcParams{
					ID:        channelRequestHtlc.ID,
					IsSettled: true,
				})

				unsettledChannelRequestHtlcs, _ := m.ChannelRequestResolver.Repository.ListUnsettledChannelRequestHtlcs(ctx, channelRequestHtlc.ChannelRequestID)
				log.Printf("Unsettled HTLCs: %v", len(unsettledChannelRequestHtlcs))

				if len(unsettledChannelRequestHtlcs) == 0 {
					// TODO: Add creating channel task to worker group
					if channelRequest, err := m.ChannelRequestResolver.Repository.GetChannelRequest(ctx, channelRequestHtlc.ChannelRequestID); err == nil {
						if pubkeyBytes, err := hex.DecodeString(channelRequest.Pubkey); err == nil {
							pushSat := int64(lnwire.MilliSatoshi(channelRequest.AmountMsat).ToSatoshis())
							localFundingAmount := int64(float64(pushSat) * 1.25)
							openChannelRequest := &lnrpc.OpenChannelRequest{
								NodePubkey:         pubkeyBytes,
								LocalFundingAmount: localFundingAmount,
								PushSat:            pushSat,
								TargetConf:         1,
								MinConfs:           0,
								Private:            true,
								SpendUnconfirmed:   true,
							}

							log.Printf("Opening channel to %v for %v sats", channelRequest.Pubkey, pushSat)

							channelPoint, err := m.LightningService.OpenChannelSync(openChannelRequest)
							util.LogOnError("LSP005", "Error opening channel", err)

							if err == nil {
								updateChannelRequestParams := param.NewUpdateChannelRequestParams(channelRequest)
								updateChannelRequestParams.FundingTxID = channelPoint.GetFundingTxidBytes()
								updateChannelRequestParams.OutputIndex = util.SqlNullInt64(channelPoint.OutputIndex)

								m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)
							}
						}
					}
				}
			}
		}
	}
}

func (m *HtlcEventMonitor) subscribeHtlcEventInterceptions(htlcEventChan chan<- routerrpc.HtlcEvent) {
	htlcEventsClient, err := m.waitForSubscribeHtlcEventsClient(0, 1000)
	util.PanicOnError("LSP018", "Error creating Htlc Events client", err)
	m.HtlcEventsClient = htlcEventsClient

	for {
		htlcInterceptRequest, err := m.HtlcEventsClient.Recv()

		if err == nil {
			htlcEventChan <- *htlcInterceptRequest
		} else {
			m.HtlcEventsClient, err = m.waitForSubscribeHtlcEventsClient(100, 1000)
			util.PanicOnError("LSP019", "Error creating Htlc Events client", err)
		}
	}
}

func (m *HtlcEventMonitor) waitForHtlcEvents(ctx context.Context, waitGroup *sync.WaitGroup, htlcEventChan chan routerrpc.HtlcEvent) {
	waitGroup.Add(1)
	defer close(htlcEventChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Htlc Events")
			return
		case htlcInterceptRequest := <-htlcEventChan:
			m.handleHtlcEvent(htlcInterceptRequest)
		}
	}
}

func (m *HtlcEventMonitor) waitForSubscribeHtlcEventsClient(initialDelay, retryDelay time.Duration) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeHtlcEventsClient, err := m.LightningService.SubscribeHtlcEvents(&routerrpc.SubscribeHtlcEventsRequest{})

		if err == nil {
			return subscribeHtlcEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeHtlcEvents client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
