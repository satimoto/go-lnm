package lightningnetwork

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LightningNetwork interface {
	AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error)
	GetInfo(in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error)
	HtlcInterceptor(opts ...grpc.CallOption) (routerrpc.Router_HtlcInterceptorClient, error)
	OpenChannelSync(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error)
	RegisterBlockEpochNtfn(in *chainrpc.BlockEpoch, opts ...grpc.CallOption) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error)
	SendCustomMessage(in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error)
	SubscribeChannelBackups(in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error)
	SubscribeChannelEvents(in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error)
	SubscribeCustomMessages(in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error)
	SubscribeHtlcEvents(in *routerrpc.SubscribeHtlcEventsRequest, opts ...grpc.CallOption) (routerrpc.Router_SubscribeHtlcEventsClient, error)
	SubscribeInvoices(in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error)
	SubscribeTransactions(in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error)
	WalletBalance(in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error)
}

type LightningNetworkService struct {
	clientConn          *grpc.ClientConn
	chainNotifierClient *chainrpc.ChainNotifierClient
	lightningClient     *lnrpc.LightningClient
	routerClient        *routerrpc.RouterClient
	macaroonCtx         context.Context
}

func NewService() LightningNetwork {
	lndTlsCert, err := base64.StdEncoding.DecodeString(os.Getenv("LND_TLS_CERT"))
	dbUtil.PanicOnError("LSP006", "Invalid LND TLS Certificate", err)

	credentials, err := util.NewCredential(string(lndTlsCert))
	dbUtil.PanicOnError("LSP007", "Error creating transport credentials", err)

	clientConn, err := grpc.Dial(os.Getenv("LND_GRPC_HOST"), grpc.WithTransportCredentials(credentials))
	dbUtil.PanicOnError("LSP008", "Error connecting to LND host", err)

	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	dbUtil.PanicOnError("LSP009", "Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))

	return &LightningNetworkService{
		clientConn:  clientConn,
		macaroonCtx: macaroonCtx,
	}
}

func (s *LightningNetworkService) AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	return s.getLightningClient().AddInvoice(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) GetInfo(in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	return s.getLightningClient().GetInfo(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) HtlcInterceptor(opts ...grpc.CallOption) (routerrpc.Router_HtlcInterceptorClient, error) {
	return s.getRouterClient().HtlcInterceptor(s.macaroonCtx, opts...)
}

func (s *LightningNetworkService) OpenChannelSync(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error) {
	return s.getLightningClient().OpenChannelSync(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) RegisterBlockEpochNtfn(in *chainrpc.BlockEpoch, opts ...grpc.CallOption) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error) {
	return s.getChainNotifierClient().RegisterBlockEpochNtfn(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SendCustomMessage(in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error) {
	return s.getLightningClient().SendCustomMessage(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeChannelBackups(in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
	return s.getLightningClient().SubscribeChannelBackups(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeChannelEvents(in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	return s.getLightningClient().SubscribeChannelEvents(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeCustomMessages(in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	return s.getLightningClient().SubscribeCustomMessages(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeHtlcEvents(in *routerrpc.SubscribeHtlcEventsRequest, opts ...grpc.CallOption) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	return s.getRouterClient().SubscribeHtlcEvents(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeInvoices(in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	return s.getLightningClient().SubscribeInvoices(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) SubscribeTransactions(in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	return s.getLightningClient().SubscribeTransactions(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) WalletBalance(in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	return s.getLightningClient().WalletBalance(s.macaroonCtx, in, opts...)
}

func (s *LightningNetworkService) getChainNotifierClient() chainrpc.ChainNotifierClient {
	if s.chainNotifierClient == nil {
		lc := chainrpc.NewChainNotifierClient(s.clientConn)
		s.chainNotifierClient = &lc
	}

	return *s.chainNotifierClient
}

func (s *LightningNetworkService) getLightningClient() lnrpc.LightningClient {
	if s.lightningClient == nil {
		lc := lnrpc.NewLightningClient(s.clientConn)
		s.lightningClient = &lc
	}

	return *s.lightningClient
}

func (s *LightningNetworkService) getRouterClient() routerrpc.RouterClient {
	if s.routerClient == nil {
		rc := routerrpc.NewRouterClient(s.clientConn)
		s.routerClient = &rc
	}

	return *s.routerClient
}

func (s *LightningNetworkService) getMacaroonCtx() context.Context {
	return s.macaroonCtx
}