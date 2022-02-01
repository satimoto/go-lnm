package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	mocks "github.com/satimoto/go-datastore-mocks/db"
	channelrequestMocks "github.com/satimoto/go-lsp/channelrequest/mocks"
	"github.com/satimoto/go-lsp/intercept"
	"google.golang.org/grpc"
)

func NewInterceptor(repositoryService *mocks.MockRepositoryService, clientConn *grpc.ClientConn) intercept.Interceptor {
	lightningClient := lnrpc.NewLightningClient(clientConn)
	routerClient := routerrpc.NewRouterClient(clientConn)
	chainNotifierClient := chainrpc.NewChainNotifierClient(clientConn)

	return &intercept.Intercept{
		ChannelRequestResolver: channelrequestMocks.NewResolver(repositoryService),
		LightningClient:        lightningClient,
		RouterClient:           routerClient,
		ChainNotifierClient:    chainNotifierClient,
	}
}