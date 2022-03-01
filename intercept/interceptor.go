package intercept

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/channelrequest"
	"github.com/satimoto/go-lsp/node"
	"google.golang.org/grpc"
)

type Interceptor interface {
	GetLightningClient() lnrpc.LightningClient
	GetRouterClient() routerrpc.RouterClient

	Register() error

	InterceptHtlcs()
	InterceptHtlc(ctx context.Context, htlcInterceptorClient routerrpc.Router_HtlcInterceptorClient) error

	SubscribeChannelEvents()
	SubscribeChannelEvent(subscribeChannelEventsClient lnrpc.Lightning_SubscribeChannelEventsClient) error

	SubscribeHtlcEvents()
	SubscribeHtlcEvent(subscribeHtlcEventsClient routerrpc.Router_SubscribeHtlcEventsClient) error

	SubscribeTransactions()
	SubscribeTransaction(subscribeTransactionsClient lnrpc.Lightning_SubscribeTransactionsClient) error
}

type Intercept struct {
	*channelrequest.ChannelRequestResolver
	*node.NodeResolver

	lnrpc.LightningClient
	routerrpc.RouterClient
	chainrpc.ChainNotifierClient

	customMessageHandlers map[string]customMessageHandler
}

func NewInterceptor(repositoryService *db.RepositoryService, clientConn *grpc.ClientConn) Interceptor {
	lightningClient := lnrpc.NewLightningClient(clientConn)
	routerClient := routerrpc.NewRouterClient(clientConn)
	chainNotifierClient := chainrpc.NewChainNotifierClient(clientConn)

	return &Intercept{
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeResolver:           node.NewResolver(repositoryService),
		LightningClient:        lightningClient,
		RouterClient:           routerClient,
		ChainNotifierClient:    chainNotifierClient,
		customMessageHandlers:  make(map[string]customMessageHandler),
	}
}

func (i *Intercept) GetLightningClient() lnrpc.LightningClient {
	return i.LightningClient
}

func (i *Intercept) GetRouterClient() routerrpc.RouterClient {
	return i.RouterClient
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

		if channelRequest.Status != db.ChannelRequestStatusAWAITINGPREIMAGE && channelRequest.Status != db.ChannelRequestStatusAWAITINGPAYMENTS {
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
