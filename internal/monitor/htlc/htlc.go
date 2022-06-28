package htlc

import (
	"bytes"
	"context"
	"encoding/hex"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/messages"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HtlcMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	HtlcInterceptorClient  routerrpc.Router_HtlcInterceptorClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	CustomMessageMonitor   *custommessage.CustomMessageMonitor
	nodeID                 int64
}

func NewHtlcMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, customMessageMonitor *custommessage.CustomMessageMonitor) *HtlcMonitor {
	return &HtlcMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		CustomMessageMonitor:   customMessageMonitor,
	}
}

func (m *HtlcMonitor) StartMonitor(nodeID int64, ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Htlcs")
	htlcInterceptChan := make(chan routerrpc.ForwardHtlcInterceptRequest)

	m.nodeID = nodeID
	go m.waitForHtlcs(ctx, waitGroup, htlcInterceptChan)
	go m.subscribeHtlcInterceptions(htlcInterceptChan)
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
		/** Channel request registered and HTLC intercepted.
		 *  We store the incoming HTLC so we can manage a failure state.
		 *  When the channel request is in a REQUESTED state, we start a payment timeout to handle cleanup of the failure state.
		 *  Add the received payment amount to total channel request amount to calculate if the payment is complete.
		 *  When the payment is complete, we settle all stored HTLCs with the provided preimage.
		 *  HTLC event subscription stream should pickup settled HTLCs. Once all are settled, the channel will be opened.
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
				m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: htlcInterceptRequest.IncomingCircuitKey,
					Action:             routerrpc.ResolveHoldForwardAction_FAIL,
				})
				return
			}

			// Store the incoming HTLC
			channelRequestHtlcParams := db.CreateChannelRequestHtlcParams{
				ChannelRequestID: channelRequest.ID,
				ChanID:           int64(htlcInterceptRequest.IncomingCircuitKey.ChanId),
				HtlcID:           int64(htlcInterceptRequest.IncomingCircuitKey.HtlcId),
				IsSettled:        false,
			}

			if _, err := m.ChannelRequestResolver.Repository.CreateChannelRequestHtlc(ctx, channelRequestHtlcParams); err != nil {
				util.LogOnError("LSP024", "Error creating channel request HTLC", err)
				m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: htlcInterceptRequest.IncomingCircuitKey,
					Action:             routerrpc.ResolveHoldForwardAction_FAIL,
				})
				return
			}

			startPaymentMonitor := channelRequest.Status == db.ChannelRequestStatusREQUESTED

			updateChannelRequestParams := param.NewUpdateChannelRequestParams(channelRequest)
			updateChannelRequestParams.Status = db.ChannelRequestStatusAWAITINGPAYMENTS
			updateChannelRequestParams.SettledMsat = channelRequest.SettledMsat + int64(htlcInterceptRequest.IncomingAmountMsat)

			// All HTLCs received, settle the HTLCs
			if updateChannelRequestParams.SettledMsat == channelRequest.AmountMsat {
				updateChannelRequestParams.Status = db.ChannelRequestStatusAWAITINGPREIMAGE
				m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

				pubkeyBytes, _ := hex.DecodeString(channelRequest.Pubkey)

				m.CustomMessageMonitor.AddHandler(func(customMessage lnrpc.CustomMessage, index string) {
					// Received a preimage peer message from pubkey peer
					pubkeyStr := hex.EncodeToString(customMessage.Peer)
					dataStr := hex.EncodeToString(customMessage.Data)

					log.Printf("Custom Message %v from %v: %v", customMessage.Type, pubkeyStr, dataStr)

					if channelRequest.Pubkey == pubkeyStr && customMessage.Type == messages.CHANNELREQUEST_RECEIVE_PREIMAGE {
						if preimage, err := lntypes.MakePreimageFromStr(dataStr); err == nil {
							log.Printf("preimage: %v", preimage.String())
							paymentHash := preimage.Hash()

							// Compare preimage hash to channel request payment hash
							if bytes.Compare(paymentHash[:], channelRequest.PaymentHash) == 0 {
								channelRequestHtlcs, _ := m.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

								m.ChannelRequestResolver.Repository.UpdateChannelRequestStatus(ctx, db.UpdateChannelRequestStatusParams{
									ID:     channelRequest.ID,
									Status: db.ChannelRequestStatusSETTLINGHTLCS,
								})

								for _, channelRequestHtlc := range channelRequestHtlcs {
									htlcInterceptResponse := &routerrpc.ForwardHtlcInterceptResponse{
										IncomingCircuitKey: &routerrpc.CircuitKey{
											ChanId: uint64(channelRequestHtlc.ChanID),
											HtlcId: uint64(channelRequestHtlc.HtlcID),
										},
										Action:   routerrpc.ResolveHoldForwardAction_SETTLE,
										Preimage: preimage[:],
									}

									m.HtlcInterceptorClient.Send(htlcInterceptResponse)
								}

								m.CustomMessageMonitor.RemoveHandler(index)
							}
						}
					}
				})

				// TODO: Ensure peer in online
				m.LightningService.SendCustomMessage(&lnrpc.SendCustomMessageRequest{
					Peer: pubkeyBytes,
					Type: messages.CHANNELREQUEST_SEND_CHAN_ID,
					Data: []byte(strconv.FormatUint(htlcInterceptRequest.OutgoingRequestedChanId, 10)),
				})
			}

			m.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

			// Start payment timeout to cleanup failures
			if startPaymentMonitor {
				// TODO: Add monitoring task to worker group, this should prevent shutdown while awaiting payments
				go m.waitForPaymentTimeout(ctx, channelRequest.PaymentHash, 30)
			}
		} else {
			log.Printf("LSP025: Invalid channel request state")
			log.Printf("LSP025: Status=%v, PaymentHash=%v", channelRequest.Status, hex.EncodeToString(htlcInterceptRequest.PaymentHash))
			m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
				IncomingCircuitKey: htlcInterceptRequest.IncomingCircuitKey,
				Action:             routerrpc.ResolveHoldForwardAction_FAIL,
			})
		}
	} else {
		m.HtlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
			IncomingCircuitKey: htlcInterceptRequest.IncomingCircuitKey,
			Action:             routerrpc.ResolveHoldForwardAction_RESUME,
		})
	}
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
			log.Printf("LSP026: Error getting channel request")
			log.Printf("LSP026: %v", err)
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
