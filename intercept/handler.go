package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/channelrequest"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor interface {
	GetRouterClient() routerrpc.RouterClient

	InterceptHtlcs()
	InterceptHtlc(htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient) error

	MonitorHtlcEvents()
	MonitorHtlcEvent(subscribeHtlcEventsClient routerrpc.Router_SubscribeHtlcEventsClient) error
}

type Intercept struct {
	*channelrequest.ChannelRequestResolver

	lnrpc.LightningClient
	routerrpc.RouterClient
	chainrpc.ChainNotifierClient
}

func NewInterceptor(repositoryService *db.RepositoryService, clientConn *grpc.ClientConn) Interceptor {
	lightningClient := lnrpc.NewLightningClient(clientConn)
	routerClient := routerrpc.NewRouterClient(clientConn)
	chainNotifierClient := chainrpc.NewChainNotifierClient(clientConn)

	return &Intercept{
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		LightningClient:        lightningClient,
		RouterClient:           routerClient,
		ChainNotifierClient:    chainNotifierClient,
	}
}

func (i *Intercept) GetRouterClient() routerrpc.RouterClient {
	return i.RouterClient
}

func (i *Intercept) InterceptHtlcs() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	htlcInterceptorClient, err := i.waitForHtlcInterceptor(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating HTLC Interceptor", err)

	for {
		if err := i.InterceptHtlc(htlcInterceptorClient); err != nil {
			htlcInterceptorClient, err = i.waitForHtlcInterceptor(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating HTLC Interceptor", err)
		}
	}
}

func (i *Intercept) InterceptHtlc(htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient) error {
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

			if _, err := i.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err != nil {
				forwardHtlcInterceptFail("Channel request HTLC already exists", err, htlcInterceptorClient, forwardHtlcInterceptRequest.IncomingCircuitKey)

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
				forwardHtlcInterceptFail("Error creating channel request HTLC", err, htlcInterceptorClient, forwardHtlcInterceptRequest.IncomingCircuitKey)

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
				updateChannelRequestParams.Status = db.ChannelRequestStatusSETTLINGHTLCS
				channelRequestHtlcs, _ := i.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

				for _, channelRequestHtlc := range channelRequestHtlcs {
					htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
						IncomingCircuitKey: &routerrpc.CircuitKey{
							ChanId: uint64(channelRequestHtlc.ChanID),
							HtlcId: uint64(channelRequestHtlc.HtlcID),
						},
						Action:   routerrpc.ResolveHoldForwardAction_SETTLE,
						Preimage: channelRequest.Preimage,
					})
				}
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

func (i *Intercept) MonitorHtlcEvents() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	subscribeHtlcEventsClient, err := i.waitForSubscribeHtlcEvents(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating client", err)

	for {
		if err := i.MonitorHtlcEvent(subscribeHtlcEventsClient); err != nil {
			subscribeHtlcEventsClient, err = i.waitForSubscribeHtlcEvents(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating client", err)
		}

	}
}

func (i *Intercept) MonitorHtlcEvent(subscribeHtlcEventsClient routerrpc.Router_SubscribeHtlcEventsClient) error {
	htlcEvent, err := subscribeHtlcEventsClient.Recv()

	if err != nil {
		log.Printf("Error receiving HTLC event: %v", status.Code(err))

		return err
	}

	/** HTLC Event received.
	 *  Check that the event type is a Forward event and that is is successful.
	 *  Find the Channel Request HTLC by the circuit key params.
	 *  If the Channel Request HTLC exists, set it as settled.
	 */

	if htlcEvent.EventType == routerrpc.HtlcEvent_FORWARD {
		successEvent := htlcEvent.GetSettleEvent()

		if successEvent != nil {
			ctx := context.Background()
			getChannelRequestHtlcByCircuitKeyParams := db.GetChannelRequestHtlcByCircuitKeyParams{
				ChanID: int64(htlcEvent.IncomingHtlcId),
				HtlcID: int64(htlcEvent.IncomingHtlcId),
			}

			if channelRequestHtlc, err := i.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err != nil {
				i.ChannelRequestResolver.Repository.UpdateChannelRequestHtlc(ctx, db.UpdateChannelRequestHtlcParams{
					ID:        channelRequestHtlc.ID,
					IsSettled: true,
				})

				unsettledChannelRequestHtlcs, _ := i.ChannelRequestResolver.Repository.ListUnsettledChannelRequestHtlcs(ctx, channelRequestHtlc.ChannelRequestID)

				if len(unsettledChannelRequestHtlcs) == 0 {
					// TODO: Add creating channel task to worker group

				}
			}
		}
	}

	return nil
}

func forwardHtlcInterceptFail(message string, err error, htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient, circuitKey *routerrpc.CircuitKey) {
	util.LogOnError(message, err)
	htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
		IncomingCircuitKey: circuitKey,
		Action:             routerrpc.ResolveHoldForwardAction_FAIL,
	})
}

func (i *Intercept) monitorPaymentTimeout(ctx context.Context, htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient, paymentHash []byte, timeoutSeconds int) {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	paymentHashString := hex.EncodeToString(paymentHash)
	log.Printf("Payment timeout set for %v seconds: %v", timeoutSeconds, paymentHashString)

	for {
		channelRequest, err := i.ChannelRequestResolver.Repository.GetChannelRequestByPaymentHash(ctx, paymentHash)

		if err != nil {
			log.Printf("Error getting channel request: %v", err)
			break
		}

		if channelRequest.Status != db.ChannelRequestStatusAWAITINGPAYMENTS {
			log.Printf("Payment timeout ended (%v): %v", channelRequest.Status, paymentHashString)
			break
		}

		if time.Now().After(deadline) {
			log.Printf("Payment timeout expired: %v", paymentHashString)

			// Update the channel request
			updateChannelRequestParams := channelrequest.NewUpdateChannelRequestParams(channelRequest)
			updateChannelRequestParams.Status = db.ChannelRequestStatusFAILED

			i.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)

			// Fail all intercepted HTLCs
			channelRequestHtlcs, _ := i.ChannelRequestResolver.Repository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

			for _, channelRequestHtlc := range channelRequestHtlcs {
				htlcInterceptorClient.Send(&routerrpc.ForwardHtlcInterceptResponse{
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

func (i *Intercept) waitForSubscribeHtlcEvents(ctx context.Context, initialDelay, retryDelay time.Duration) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeHtlcEventsClient, err := i.RouterClient.SubscribeHtlcEvents(ctx, &routerrpc.SubscribeHtlcEventsRequest{})

		if err == nil {
			return subscribeHtlcEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for subscribe HTLC events client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}

func (i *Intercept) waitForHtlcInterceptor(ctx context.Context, initialDelay, retryDelay time.Duration) (routerrpc.Router_HtlcInterceptorClient, error) {
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

		log.Print("Waiting for HTLC interceptor client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}