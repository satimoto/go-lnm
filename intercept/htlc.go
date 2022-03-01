package intercept

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/channelrequest"
	"github.com/satimoto/go-lsp/messages"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (i *Intercept) InterceptHtlcs() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	htlcInterceptorClient, err := i.waitForHtlcInterceptorClient(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating HtlcInterceptor client", err)

	for {
		if err := i.InterceptHtlc(macaroonCtx, htlcInterceptorClient); err != nil {
			htlcInterceptorClient, err = i.waitForHtlcInterceptorClient(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating HtlcInterceptor client", err)
		}
	}
}

func (i *Intercept) InterceptHtlc(macaroonCtx context.Context, htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient) error {
	forwardHtlcInterceptRequest, err := htlcInterceptorClient.Recv()

	if err != nil {
		log.Printf("Error receiving HTLC intercept: %v", status.Code(err))

		return err
	}

	log.Print("HTLC Intercept")
	log.Printf("IncomingAmountMsat: %v", forwardHtlcInterceptRequest.IncomingAmountMsat)
	log.Printf("IncomingCircuitKey.ChanId: %v", forwardHtlcInterceptRequest.IncomingCircuitKey.ChanId)
	log.Printf("IncomingCircuitKey.HtlcId: %v", forwardHtlcInterceptRequest.IncomingCircuitKey.HtlcId)
	log.Printf("IncomingExpiry: %v", forwardHtlcInterceptRequest.IncomingExpiry)
	log.Printf("OnionBlob: %v", hex.EncodeToString(forwardHtlcInterceptRequest.OnionBlob))
	log.Printf("OutgoingAmountMsat: %v", forwardHtlcInterceptRequest.OutgoingAmountMsat)
	log.Printf("OutgoingExpiry: %v", forwardHtlcInterceptRequest.OutgoingExpiry)
	log.Printf("OutgoingRequestedChanId: %v", forwardHtlcInterceptRequest.OutgoingRequestedChanId)
	log.Printf("PaymentHash: %v", hex.EncodeToString(forwardHtlcInterceptRequest.PaymentHash))

	ctx := context.Background()
	channelRequest, err := i.ChannelRequestResolver.Repository.GetChannelRequestByPaymentHash(ctx, forwardHtlcInterceptRequest.PaymentHash)

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
				ChanID: int64(forwardHtlcInterceptRequest.IncomingCircuitKey.ChanId),
				HtlcID: int64(forwardHtlcInterceptRequest.IncomingCircuitKey.HtlcId),
			}

			if _, err := i.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err == nil {
				log.Printf("Channel request HTLC already exists (%v, %v): %v",
					forwardHtlcInterceptRequest.IncomingCircuitKey.ChanId,
					forwardHtlcInterceptRequest.IncomingCircuitKey.HtlcId,
					hex.EncodeToString(forwardHtlcInterceptRequest.PaymentHash))
				htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: forwardHtlcInterceptRequest.IncomingCircuitKey,
					Action:             routerrpc.ResolveHoldForwardAction_FAIL,
				})
				return nil
			}

			// Store the incoming HTLC
			channelRequestHtlcParams := db.CreateChannelRequestHtlcParams{
				ChannelRequestID: channelRequest.ID,
				ChanID:           int64(forwardHtlcInterceptRequest.IncomingCircuitKey.ChanId),
				HtlcID:           int64(forwardHtlcInterceptRequest.IncomingCircuitKey.HtlcId),
				IsSettled:        false,
			}

			if _, err := i.ChannelRequestResolver.Repository.CreateChannelRequestHtlc(ctx, channelRequestHtlcParams); err != nil {
				log.Printf("Error creating channel request HTLC: %v", err)
				htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
					IncomingCircuitKey: forwardHtlcInterceptRequest.IncomingCircuitKey,
					Action:             routerrpc.ResolveHoldForwardAction_FAIL,
				})
				return nil
			}

			// Start payment timeout to cleanup failures
			if channelRequest.Status == db.ChannelRequestStatusREQUESTED {
				// TODO: Add monitoring task to worker group, this should prevent shutdown while awaiting payments
				go i.monitorPaymentTimeout(ctx, htlcInterceptorClient, channelRequest.PaymentHash, 30)
			}

			updateChannelRequestParams := channelrequest.NewUpdateChannelRequestParams(channelRequest)
			updateChannelRequestParams.Status = db.ChannelRequestStatusAWAITINGPAYMENTS
			updateChannelRequestParams.SettledMsat = channelRequest.SettledMsat + int64(forwardHtlcInterceptRequest.IncomingAmountMsat)

			// All HTLCs received, settle the HTLCs
			if updateChannelRequestParams.SettledMsat == channelRequest.AmountMsat {
				updateChannelRequestParams.Status = db.ChannelRequestStatusAWAITINGPREIMAGE
				i.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

				pubkeyBytes, _ := hex.DecodeString(channelRequest.Pubkey)

				// TODO: Ensure peer in online
				i.LightningClient.SendCustomMessage(macaroonCtx, &lnrpc.SendCustomMessageRequest{
					Peer: pubkeyBytes,
					Type: messages.CHANNELREQUEST_SEND_CHAN_ID,
					Data: []byte(strconv.FormatUint(forwardHtlcInterceptRequest.OutgoingRequestedChanId, 10)),
				})

				i.AddCustomMessageHandler(func(customMessage *lnrpc.CustomMessage, index string) {
					// Received a preimage peer message from pubkey peer
					if bytes.Compare(customMessage.Peer, pubkeyBytes) == 0 && customMessage.Type == messages.CHANNELREQUEST_RECEIVE_PREIMAGE {
						if preimage, err := lntypes.MakePreimageFromStr(string(customMessage.Data)); err == nil {
							paymentHash := preimage.Hash()

							// Compare preimage hash to channel request payment hash
							if bytes.Compare(paymentHash[:], channelRequest.PaymentHash) == 0 {
								channelRequestHtlcs, _ := i.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

								i.ChannelRequestResolver.Repository.UpdateChannelRequestStatus(ctx, db.UpdateChannelRequestStatusParams{
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

									htlcInterceptorClient.Send(htlcInterceptResponse)
								}

								i.RemoveCustomMessageHandler(index)
							}
						}
					}
				})
			}

			i.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)
		} else {
			log.Printf("Invalid channel request state (%v): %v", channelRequest.Status, hex.EncodeToString(forwardHtlcInterceptRequest.PaymentHash))

			htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
				IncomingCircuitKey: forwardHtlcInterceptRequest.IncomingCircuitKey,
				Action:             routerrpc.ResolveHoldForwardAction_FAIL,
			})
		}
	} else {
		htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
			IncomingCircuitKey: forwardHtlcInterceptRequest.IncomingCircuitKey,
			Action:             routerrpc.ResolveHoldForwardAction_RESUME,
		})
	}

	return nil
}

func (i *Intercept) waitForHtlcInterceptorClient(ctx context.Context, initialDelay, retryDelay time.Duration) (routerrpc.Router_HtlcInterceptorClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		htlcInterceptorClient, err := i.RouterClient.HtlcInterceptor(ctx)

		if err == nil {
			return htlcInterceptorClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for HtlcInterceptor client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
