package htlc

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/monitor/psbtfund"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HtlcMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	PsbtFundService        psbtfund.PsbtFund
	HtlcInterceptorClient  routerrpc.Router_HtlcInterceptorClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	nodeID                 int64
}

func NewHtlcMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, psbtFundService psbtfund.PsbtFund) *HtlcMonitor {
	return &HtlcMonitor{
		LightningService:       lightningService,
		PsbtFundService:        psbtFundService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}

func (m *HtlcMonitor) StartMonitor(nodeID int64, ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Htlcs")
	htlcInterceptChan := make(chan routerrpc.ForwardHtlcInterceptRequest)

	m.nodeID = nodeID
	go m.waitForHtlcs(ctx, waitGroup, htlcInterceptChan)
	go m.subscribeHtlcInterceptions(htlcInterceptChan)
}

func (m *HtlcMonitor) ResumeChannelRequestHtlcs(channelRequest db.ChannelRequest) {
	if channelRequest.Status == db.ChannelRequestStatusOPENINGCHANNEL {
		log.Printf("Wait 10 seconds before resuming HTLCs: %v", channelRequest.ID)
		time.Sleep(10 * time.Second)

		ctx := context.Background()
		channelRequestHtlcs, err := m.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

		if err != nil {
			util.LogOnError("LSP097", "Error listing channel request HTLCs", err)
			log.Printf("LSP097: ChannelRequestID=%v", channelRequest.ID)
		}

		for _, channelRequestHtlc := range channelRequestHtlcs {
			if !channelRequestHtlc.IsSettled && !channelRequestHtlc.IsFailed {
				htlcInterceptResponse := &routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: &routerrpc.CircuitKey{
						ChanId: uint64(channelRequestHtlc.ChanID),
						HtlcId: uint64(channelRequestHtlc.HtlcID),
					},
					Action: routerrpc.ResolveHoldForwardAction_RESUME,
				}
	
				m.HtlcInterceptorClient.Send(htlcInterceptResponse)
			}
		}

		updateChannelRequestParams := param.NewUpdateChannelRequestParams(channelRequest)
		updateChannelRequestParams.Status = db.ChannelRequestStatusCOMPLETED

		_, err = m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

		if err != nil {
			util.LogOnError("LSP099", "Error updating channel request", err)
			log.Printf("LSP099: Params=%#v", updateChannelRequestParams)
		}
	}
}

func (m *HtlcMonitor) handleHtlc(htlcInterceptRequest routerrpc.ForwardHtlcInterceptRequest) {
	log.Print("HTLC Intercept")
	log.Printf("IncomingAmountMsat: %v", htlcInterceptRequest.IncomingAmountMsat)
	log.Printf("IncomingCircuitKey.ChanId: %v", htlcInterceptRequest.IncomingCircuitKey.ChanId)
	log.Printf("IncomingCircuitKey.HtlcId: %v", htlcInterceptRequest.IncomingCircuitKey.HtlcId)
	log.Printf("IncomingExpiry: %v", htlcInterceptRequest.IncomingExpiry)
	log.Printf("OnionBlob: %v", hex.EncodeToString(htlcInterceptRequest.OnionBlob))
	log.Printf("OutgoingAmountMsat: %v", htlcInterceptRequest.OutgoingAmountMsat)
	log.Printf("OutgoingExpiry: %v", htlcInterceptRequest.OutgoingExpiry)
	log.Printf("OutgoingRequestedChanId: %v", htlcInterceptRequest.OutgoingRequestedChanId)
	log.Printf("PaymentHash: %v", hex.EncodeToString(htlcInterceptRequest.PaymentHash))

	ctx := context.Background()
	channelRequest, err := m.ChannelRequestResolver.Repository.GetChannelRequestByPaymentHash(ctx, htlcInterceptRequest.PaymentHash)

	if err == nil {
		log.Printf("Channel request registered and HTLC intercepted")
		/** Channel request registered and HTLC intercepted.
		 *  We store the incoming HTLC so we can manage a failure state.
		 *  When the channel request is in a REQUESTED state, we start a payment timeout to handle cleanup of the failure state.
		 *  Add the received payment amount to total channel request amount to calculate if the payment is complete.
		 *  When the payment is complete, we use PsbtFundService to open   channel.
		 *  Channel event subscription stream will pick up when the channel is open and resume the payment.
		 */

		if channelRequest.Status == db.ChannelRequestStatusREQUESTED || channelRequest.Status == db.ChannelRequestStatusAWAITINGPAYMENTS {
			// Check the HTLC has not already been handled
			getChannelRequestHtlcByCircuitKeyParams := db.GetChannelRequestHtlcByCircuitKeyParams{
				ChanID: int64(htlcInterceptRequest.IncomingCircuitKey.ChanId),
				HtlcID: int64(htlcInterceptRequest.IncomingCircuitKey.HtlcId),
			}

			if _, err := m.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err == nil {
				log.Printf("Channel request HTLC already exists (%v, %v): %v",
					htlcInterceptRequest.IncomingCircuitKey.ChanId,
					htlcInterceptRequest.IncomingCircuitKey.HtlcId,
					hex.EncodeToString(htlcInterceptRequest.PaymentHash))
				m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
				return
			}

			// Store the incoming HTLC
			channelRequestHtlcParams := db.CreateChannelRequestHtlcParams{
				ChannelRequestID: channelRequest.ID,
				ChanID:           int64(htlcInterceptRequest.IncomingCircuitKey.ChanId),
				HtlcID:           int64(htlcInterceptRequest.IncomingCircuitKey.HtlcId),
				AmountMsat:       int64(htlcInterceptRequest.OutgoingAmountMsat),
				IsSettled:        false,
				IsFailed:         false,
			}

			if _, err := m.ChannelRequestResolver.Repository.CreateChannelRequestHtlc(ctx, channelRequestHtlcParams); err != nil {
				util.LogOnError("LSP024", "Error creating channel request HTLC", err)
				m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
				return
			}

			startPaymentMonitor := channelRequest.Status == db.ChannelRequestStatusREQUESTED

			updateChannelRequestParams := param.NewUpdateChannelRequestParams(channelRequest)
			updateChannelRequestParams.Status = db.ChannelRequestStatusAWAITINGPAYMENTS
			updateChannelRequestParams.SettledMsat = channelRequest.SettledMsat + int64(htlcInterceptRequest.OutgoingAmountMsat)

			// All HTLCs received, settle the HTLCs
			if updateChannelRequestParams.SettledMsat == channelRequest.AmountMsat {
				pubkeyBytes, err := hex.DecodeString(channelRequest.Pubkey)

				if err != nil {
					util.LogOnError("LSP052", "Error decoding pubkey", err)
					log.Printf("LSP052: ChannelRequestID=%#v", channelRequest.ID)
					m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
					return
				}

				// Requested channel amount plus 25%
				localFundingAmount := int64(float64(channelRequest.Amount) * 1.25)

				// TODO: Verify the are enough funds to open the channel
				openChannelRequest := &lnrpc.OpenChannelRequest{
					NodePubkey:         pubkeyBytes,
					LocalFundingAmount: localFundingAmount,
					CommitmentType:     lnrpc.CommitmentType_ANCHORS,
					Private:            true,
					ZeroConf:           true,
					ScidAlias:          true,
				}

				err = m.PsbtFundService.OpenChannel(ctx, openChannelRequest, channelRequest)

				if err != nil {
					util.LogOnError("LSP053", "Error opening channel", err)
					log.Printf("LSP053: OpenChannelRequest=%#v", openChannelRequest)
					m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
					return
				}

				updateChannelRequestParams.FundingAmount = util.SqlNullInt64(localFundingAmount)
				updateChannelRequestParams.Status = db.ChannelRequestStatusOPENINGCHANNEL
			}

			_, err := m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

			if err != nil {
				util.LogOnError("LSP096", "Error updating channel request", err)
				log.Printf("LSP096: Params=%#v", updateChannelRequestParams)
				m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
				return
			}

			// Start payment timeout to cleanup failures
			if startPaymentMonitor {
				// TODO: Add monitoring task to worker group, this should prevent shutdown while awaiting payments
				go m.waitForPaymentTimeout(ctx, channelRequest.PaymentHash, 30)
			}
		} else {
			log.Printf("LSP025: Invalid channel request state")
			log.Printf("LSP025: Status=%v, PaymentHash=%v", channelRequest.Status, hex.EncodeToString(htlcInterceptRequest.PaymentHash))
			m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_FAIL)
		}
	} else {
		util.LogOnError("LSP0X1", "Request not found, resuming HTLC", err)
		m.sendToHtlcInterceptor(htlcInterceptRequest.IncomingCircuitKey, routerrpc.ResolveHoldForwardAction_RESUME)
	}
}

func (m *HtlcMonitor) sendToHtlcInterceptor(incomingCircuitKey *routerrpc.CircuitKey, action routerrpc.ResolveHoldForwardAction) {
	m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
		IncomingCircuitKey: incomingCircuitKey,
		Action:             action,
	})
}

func (m *HtlcMonitor) subscribeHtlcInterceptions(htlcInterceptChan chan<- routerrpc.ForwardHtlcInterceptRequest) {
	htlcInterceptorClient, err := m.waitForHtlcInterceptorClient(0, 1000)
	util.PanicOnError("LSP016", "Error creating Htlcs client", err)
	m.HtlcInterceptorClient = htlcInterceptorClient

	for {
		htlcInterceptRequest, err := m.HtlcInterceptorClient.Recv()

		if err == nil {
			htlcInterceptChan <- *htlcInterceptRequest
		} else {
			m.HtlcInterceptorClient, err = m.waitForHtlcInterceptorClient(100, 1000)
			util.PanicOnError("LSP017", "Error creating Htlcs client", err)
		}
	}
}

func (m *HtlcMonitor) waitForHtlcs(ctx context.Context, waitGroup *sync.WaitGroup, htlcInterceptChan chan routerrpc.ForwardHtlcInterceptRequest) {
	waitGroup.Add(1)
	defer close(htlcInterceptChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Htlcs")
			return
		case htlcInterceptRequest := <-htlcInterceptChan:
			m.handleHtlc(htlcInterceptRequest)
		}
	}
}

func (m *HtlcMonitor) waitForHtlcInterceptorClient(initialDelay, retryDelay time.Duration) (routerrpc.Router_HtlcInterceptorClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		htlcInterceptorClient, err := m.LightningService.HtlcInterceptor()

		if err == nil {
			return htlcInterceptorClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Htlcs client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}

func (m *HtlcMonitor) waitForPaymentTimeout(ctx context.Context, paymentHash []byte, timeoutSeconds int) {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	paymentHashString := hex.EncodeToString(paymentHash)
	log.Printf("Payment timeout set for %v seconds: %v", timeoutSeconds, paymentHashString)

	for {
		channelRequest, err := m.ChannelRequestResolver.Repository.GetChannelRequestByPaymentHash(ctx, paymentHash)

		if err != nil {
			util.LogOnError("LSP026", "Error getting channel request", err)
			break
		}

		if channelRequest.Status != db.ChannelRequestStatusAWAITINGPREIMAGE && channelRequest.Status != db.ChannelRequestStatusAWAITINGPAYMENTS {
			log.Printf("Payment timeout ended (%v): %v", channelRequest.Status, paymentHashString)
			break
		}

		if time.Now().After(deadline) {
			log.Printf("Payment timeout expired: %v", paymentHashString)

			// Update the channel request
			updateChannelRequestParams := param.NewUpdateChannelRequestParams(channelRequest)
			updateChannelRequestParams.Status = db.ChannelRequestStatusFAILED

			m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

			// Fail all intercepted HTLCs
			channelRequestHtlcs, _ := m.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

			for _, channelRequestHtlc := range channelRequestHtlcs {
				m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: &routerrpc.CircuitKey{
						ChanId: uint64(channelRequestHtlc.ChanID),
						HtlcId: uint64(channelRequestHtlc.HtlcID),
					},
					Action: routerrpc.ResolveHoldForwardAction_FAIL,
				})
			}

			break
		}

		time.Sleep(1 * time.Second)
	}
}
