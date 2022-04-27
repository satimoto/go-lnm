package lightningnetwork

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LightningNetwork interface {
	GetLightningClient() lnrpc.LightningClient
	GetRouterClient() routerrpc.RouterClient
	GetMacaroonCtx() context.Context
}

type LightningNetworkService struct {
	clientConn      *grpc.ClientConn
	lightningClient *lnrpc.LightningClient
	routerClient    *routerrpc.RouterClient
	macaroonCtx     context.Context
}

func NewService() LightningNetwork {
	lndTlsCert, err := base64.StdEncoding.DecodeString(os.Getenv("LND_TLS_CERT"))
	util.PanicOnError("LSP006", "Invalid LND TLS Certificate", err)

	credentials, err := util.NewCredential(string(lndTlsCert))
	util.PanicOnError("LSP007", "Error creating transport credentials", err)

	clientConn, err := grpc.Dial(os.Getenv("LND_GRPC_HOST"), grpc.WithTransportCredentials(credentials))
	util.PanicOnError("LSP008", "Error connecting to LND host", err)

	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("LSP009", "Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))

	return &LightningNetworkService{
		clientConn:  clientConn,
		macaroonCtx: macaroonCtx,
	}
}

func (s *LightningNetworkService) GetLightningClient() lnrpc.LightningClient {
	if s.lightningClient == nil {
		lc := lnrpc.NewLightningClient(s.clientConn)
		s.lightningClient = &lc
	}

	return *s.lightningClient
}

func (s *LightningNetworkService) GetRouterClient() routerrpc.RouterClient {
	if s.routerClient == nil {
		rc := routerrpc.NewRouterClient(s.clientConn)
		s.routerClient = &rc
	}

	return *s.routerClient
}

func (s *LightningNetworkService) GetMacaroonCtx() context.Context {
	return s.macaroonCtx
}
