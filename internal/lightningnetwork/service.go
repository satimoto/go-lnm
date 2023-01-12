package lightningnetwork

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
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LightningNetwork interface {
	AllocateAlias(in *lnrpc.AllocateAliasRequest, opts ...grpc.CallOption) (*lnrpc.AllocateAliasResponse, error)
	AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error)
	ChannelAcceptor(opts ...grpc.CallOption) (lnrpc.Lightning_ChannelAcceptorClient, error)
	DecodePayReq(in *lnrpc.PayReqString, opts ...grpc.CallOption) (*lnrpc.PayReq, error)
	EstimateFee(in *walletrpc.EstimateFeeRequest, opts ...grpc.CallOption) (*walletrpc.EstimateFeeResponse, error)
	FinalizePsbt(in *walletrpc.FinalizePsbtRequest, opts ...grpc.CallOption) (*walletrpc.FinalizePsbtResponse, error)
	FundingStateStep(in *lnrpc.FundingTransitionMsg, opts ...grpc.CallOption) (*lnrpc.FundingStateStepResp, error)
	FundPsbt(in *walletrpc.FundPsbtRequest, opts ...grpc.CallOption) (*walletrpc.FundPsbtResponse, error)
	GetInfo(in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error)
	HtlcInterceptor(opts ...grpc.CallOption) (routerrpc.Router_HtlcInterceptorClient, error)
	ListChannels(in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error)
	ListPeers(in *lnrpc.ListPeersRequest, opts ...grpc.CallOption) (*lnrpc.ListPeersResponse, error)
	OpenChannel(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error)
	OpenChannelSync(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error)
	PublishTransaction(in *walletrpc.Transaction, opts ...grpc.CallOption) (*walletrpc.PublishResponse, error)
	RegisterBlockEpochNtfn(in *chainrpc.BlockEpoch, opts ...grpc.CallOption) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error)
	SendCustomMessage(in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error)
	SendPaymentV2(in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentV2Client, error)
	SignMessage(in *lnrpc.SignMessageRequest, opts ...grpc.CallOption) (*lnrpc.SignMessageResponse, error)
	SubscribeChannelBackups(in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error)
	SubscribeChannelEvents(in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error)
	SubscribeChannelGraph(in *lnrpc.GraphTopologySubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error)
	SubscribeCustomMessages(in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error)
	SubscribeHtlcEvents(in *routerrpc.SubscribeHtlcEventsRequest, opts ...grpc.CallOption) (routerrpc.Router_SubscribeHtlcEventsClient, error)
	SubscribeInvoices(in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error)
	SubscribePeerEvents(in *lnrpc.PeerEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error)
	SubscribeTransactions(in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error)
	UpdateChannelPolicy(in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error)
	WalletBalance(in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error)
}

type LightningNetworkService struct {
	clientConn          *grpc.ClientConn
	chainNotifierClient *chainrpc.ChainNotifierClient
	lightningClient     *lnrpc.LightningClient
	routerClient        *routerrpc.RouterClient
	walletKitClient     *walletrpc.WalletKitClient
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

func (s *LightningNetworkService) AllocateAlias(in *lnrpc.AllocateAliasRequest, opts ...grpc.CallOption) (*lnrpc.AllocateAliasResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().AllocateAlias(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("AllocateAlias responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().AddInvoice(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("AddInvoice responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) ChannelAcceptor(opts ...grpc.CallOption) (lnrpc.Lightning_ChannelAcceptorClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().ChannelAcceptor(s.macaroonCtx, opts...)
	timerStop := time.Now()

	log.Printf("ChannelAcceptor responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) DecodePayReq(in *lnrpc.PayReqString, opts ...grpc.CallOption) (*lnrpc.PayReq, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().DecodePayReq(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("DecodePayReq responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) EstimateFee(in *walletrpc.EstimateFeeRequest, opts ...grpc.CallOption) (*walletrpc.EstimateFeeResponse, error) {
	timerStart := time.Now()
	response, err := s.getWalletKitClient().EstimateFee(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("EstimateFee responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) FinalizePsbt(in *walletrpc.FinalizePsbtRequest, opts ...grpc.CallOption) (*walletrpc.FinalizePsbtResponse, error) {
	timerStart := time.Now()
	response, err := s.getWalletKitClient().FinalizePsbt(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("FinalizePsbt responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) FundingStateStep(in *lnrpc.FundingTransitionMsg, opts ...grpc.CallOption) (*lnrpc.FundingStateStepResp, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().FundingStateStep(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("FundingStateStep responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) FundPsbt(in *walletrpc.FundPsbtRequest, opts ...grpc.CallOption) (*walletrpc.FundPsbtResponse, error) {
	timerStart := time.Now()
	response, err := s.getWalletKitClient().FundPsbt(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("FundPsbt responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) GetInfo(in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().GetInfo(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("GetInfo responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) HtlcInterceptor(opts ...grpc.CallOption) (routerrpc.Router_HtlcInterceptorClient, error) {
	timerStart := time.Now()
	response, err := s.getRouterClient().HtlcInterceptor(s.macaroonCtx, opts...)
	timerStop := time.Now()

	log.Printf("HtlcInterceptor responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) ListChannels(in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().ListChannels(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("ListChannels responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) ListPeers(in *lnrpc.ListPeersRequest, opts ...grpc.CallOption) (*lnrpc.ListPeersResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().ListPeers(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("ListPeers responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) OpenChannel(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().OpenChannel(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("OpenChannel responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) OpenChannelSync(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().OpenChannelSync(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("OpenChannelSync responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) PublishTransaction(in *walletrpc.Transaction, opts ...grpc.CallOption) (*walletrpc.PublishResponse, error) {
	timerStart := time.Now()
	response, err := s.getWalletKitClient().PublishTransaction(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("PublishTransaction responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) RegisterBlockEpochNtfn(in *chainrpc.BlockEpoch, opts ...grpc.CallOption) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error) {
	timerStart := time.Now()
	response, err := s.getChainNotifierClient().RegisterBlockEpochNtfn(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("RegisterBlockEpochNtfn responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SendCustomMessage(in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SendCustomMessage(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SendCustomMessage responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SendPaymentV2(in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentV2Client, error) {
	timerStart := time.Now()
	response, err := s.getRouterClient().SendPaymentV2(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SendPaymentV2 responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SignMessage(in *lnrpc.SignMessageRequest, opts ...grpc.CallOption) (*lnrpc.SignMessageResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SignMessage(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SignMessage responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeChannelBackups(in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeChannelBackups(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeChannelBackups responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeChannelEvents(in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeChannelEvents(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeChannelEvents responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeChannelGraph(in *lnrpc.GraphTopologySubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeChannelGraph(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeChannelGraph responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeCustomMessages(in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeCustomMessages(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeCustomMessages responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeHtlcEvents(in *routerrpc.SubscribeHtlcEventsRequest, opts ...grpc.CallOption) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	timerStart := time.Now()
	response, err := s.getRouterClient().SubscribeHtlcEvents(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeHtlcEvents responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeInvoices(in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeInvoices(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeInvoices responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribePeerEvents(in *lnrpc.PeerEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribePeerEvents(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribePeerEvents responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) SubscribeTransactions(in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().SubscribeTransactions(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("SubscribeTransactions responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) UpdateChannelPolicy(in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().UpdateChannelPolicy(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("UpdateChannelPolicy responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LightningNetworkService) WalletBalance(in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	timerStart := time.Now()
	response, err := s.getLightningClient().WalletBalance(s.macaroonCtx, in, opts...)
	timerStop := time.Now()

	log.Printf("WalletBalance responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
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

func (s *LightningNetworkService) getWalletKitClient() walletrpc.WalletKitClient {
	if s.walletKitClient == nil {
		rc := walletrpc.NewWalletKitClient(s.clientConn)
		s.walletKitClient = &rc
	}

	return *s.walletKitClient
}

func (s *LightningNetworkService) getMacaroonCtx() context.Context {
	return s.macaroonCtx
}
