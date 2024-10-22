package mocks

import (
	"encoding/hex"
	"errors"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"google.golang.org/grpc"
)

type MockLightningNetworkService struct {
	allocateAliasMockData           []*lnrpc.AllocateAliasResponse
	addInvoiceMockData              []*lnrpc.Invoice
	channelAcceptorMockData         []lnrpc.Lightning_ChannelAcceptorClient
	decodePayReqMockData            []*lnrpc.PayReq
	estimateFeeMockData             []*walletrpc.EstimateFeeResponse
	finalizePsbtMockData            []*walletrpc.FinalizePsbtResponse
	fundingStateStepMockData        []*lnrpc.FundingStateStepResp
	fundPsbtMockData                []*walletrpc.FundPsbtResponse
	getInfoMockData                 []*lnrpc.GetInfoResponse
	htlcInterceptorMockData         []routerrpc.Router_HtlcInterceptorClient
	listChannelsMockData            []*lnrpc.ListChannelsResponse
	listPeersMockData               []*lnrpc.ListPeersResponse
	openChannelMockData             []lnrpc.Lightning_OpenChannelClient
	openChannelSyncMockData         []*lnrpc.ChannelPoint
	publishTransactionMockData      []*walletrpc.PublishResponse
	registerBlockEpochNtfnMockData  []chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient
	sendCustomMessageMockData       []*lnrpc.SendCustomMessageResponse
	sendPaymentV2MockData           []routerrpc.Router_SendPaymentV2Client
	signMessageMockData             []*lnrpc.SignMessageResponse
	subscribeChannelBackupsMockData []lnrpc.Lightning_SubscribeChannelBackupsClient
	subscribeChannelEventsMockData  []lnrpc.Lightning_SubscribeChannelEventsClient
	subscribeChannelGraphMockData   []lnrpc.Lightning_SubscribeChannelGraphClient
	subscribeCustomMessagesMockData []lnrpc.Lightning_SubscribeCustomMessagesClient
	subscribeHtlcEventsMockData     []routerrpc.Router_SubscribeHtlcEventsClient
	subscribeInvoicesMockData       []lnrpc.Lightning_SubscribeInvoicesClient
	subscribePeerEventsMockData     []lnrpc.Lightning_SubscribePeerEventsClient
	subscribeTransactionsMockData   []lnrpc.Lightning_SubscribeTransactionsClient
	updateChannelPolicyMockData     []*lnrpc.PolicyUpdateResponse
	walletBalanceMockData           []*lnrpc.WalletBalanceResponse
}

func NewService() *MockLightningNetworkService {
	return &MockLightningNetworkService{}
}

func (s *MockLightningNetworkService) AllocateAlias(in *lnrpc.AllocateAliasRequest, opts ...grpc.CallOption) (*lnrpc.AllocateAliasResponse, error) {
	if len(s.allocateAliasMockData) == 0 {
		return &lnrpc.AllocateAliasResponse{}, errors.New("NotFound")
	}

	response := s.allocateAliasMockData[0]
	s.allocateAliasMockData = s.allocateAliasMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetAllocateAliasMockData(mockData *lnrpc.AllocateAliasResponse) {
	s.allocateAliasMockData = append(s.allocateAliasMockData, mockData)
}

func (s *MockLightningNetworkService) AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	s.addInvoiceMockData = append(s.addInvoiceMockData, in)

	return &lnrpc.AddInvoiceResponse{
		RHash:          in.RHash,
		PaymentRequest: hex.EncodeToString(in.RPreimage),
	}, nil
}

func (s *MockLightningNetworkService) ChannelAcceptor(opts ...grpc.CallOption) (lnrpc.Lightning_ChannelAcceptorClient, error) {
	if len(s.channelAcceptorMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.channelAcceptorMockData[0]
	s.channelAcceptorMockData = s.channelAcceptorMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewChannelAcceptorMockData() (<-chan *lnrpc.ChannelAcceptResponse, chan<- *lnrpc.ChannelAcceptRequest) {
	sendChan := make(chan *lnrpc.ChannelAcceptResponse)
	recvChan := make(chan *lnrpc.ChannelAcceptRequest)
	s.channelAcceptorMockData = append(s.channelAcceptorMockData, NewMockChannelAcceptorClient(sendChan, recvChan))

	return sendChan, recvChan
}

func (s *MockLightningNetworkService) GetAddInvoiceMockData() (*lnrpc.Invoice, error) {
	if len(s.addInvoiceMockData) == 0 {
		return &lnrpc.Invoice{}, errors.New("NotFound")
	}

	response := s.addInvoiceMockData[0]
	s.addInvoiceMockData = s.addInvoiceMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) DecodePayReq(in *lnrpc.PayReqString, opts ...grpc.CallOption) (*lnrpc.PayReq, error) {
	if len(s.decodePayReqMockData) == 0 {
		return &lnrpc.PayReq{}, errors.New("NotFound")
	}

	response := s.decodePayReqMockData[0]
	s.decodePayReqMockData = s.decodePayReqMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetDecodePayReqMockData(mockData *lnrpc.PayReq) {
	s.decodePayReqMockData = append(s.decodePayReqMockData, mockData)
}

func (s *MockLightningNetworkService) EstimateFee(in *walletrpc.EstimateFeeRequest, opts ...grpc.CallOption) (*walletrpc.EstimateFeeResponse, error) {
	if len(s.estimateFeeMockData) == 0 {
		return &walletrpc.EstimateFeeResponse{}, errors.New("NotFound")
	}

	response := s.estimateFeeMockData[0]
	s.estimateFeeMockData = s.estimateFeeMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetEstimateFeeMockData(mockData *walletrpc.EstimateFeeResponse) {
	s.estimateFeeMockData = append(s.estimateFeeMockData, mockData)
}

func (s *MockLightningNetworkService) FinalizePsbt(in *walletrpc.FinalizePsbtRequest, opts ...grpc.CallOption) (*walletrpc.FinalizePsbtResponse, error) {
	if len(s.finalizePsbtMockData) == 0 {
		return &walletrpc.FinalizePsbtResponse{}, errors.New("NotFound")
	}

	response := s.finalizePsbtMockData[0]
	s.finalizePsbtMockData = s.finalizePsbtMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetFinalizePsbtMockData(mockData *walletrpc.FinalizePsbtResponse) {
	s.finalizePsbtMockData = append(s.finalizePsbtMockData, mockData)
}

func (s *MockLightningNetworkService) FundingStateStep(in *lnrpc.FundingTransitionMsg, opts ...grpc.CallOption) (*lnrpc.FundingStateStepResp, error) {
	if len(s.fundingStateStepMockData) == 0 {
		return &lnrpc.FundingStateStepResp{}, errors.New("NotFound")
	}

	response := s.fundingStateStepMockData[0]
	s.fundingStateStepMockData = s.fundingStateStepMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetFundingStateStepMockData(mockData *lnrpc.FundingStateStepResp) {
	s.fundingStateStepMockData = append(s.fundingStateStepMockData, mockData)
}

func (s *MockLightningNetworkService) FundPsbt(in *walletrpc.FundPsbtRequest, opts ...grpc.CallOption) (*walletrpc.FundPsbtResponse, error) {
	if len(s.fundPsbtMockData) == 0 {
		return &walletrpc.FundPsbtResponse{}, errors.New("NotFound")
	}

	response := s.fundPsbtMockData[0]
	s.fundPsbtMockData = s.fundPsbtMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetFundPsbtMockData(mockData *walletrpc.FundPsbtResponse) {
	s.fundPsbtMockData = append(s.fundPsbtMockData, mockData)
}

func (s *MockLightningNetworkService) GetInfo(in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	if len(s.getInfoMockData) == 0 {
		return &lnrpc.GetInfoResponse{}, errors.New("NotFound")
	}

	response := s.getInfoMockData[0]
	s.getInfoMockData = s.getInfoMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetGetInfoMockData(mockData *lnrpc.GetInfoResponse) {
	s.getInfoMockData = append(s.getInfoMockData, mockData)
}

func (s *MockLightningNetworkService) HtlcInterceptor(opts ...grpc.CallOption) (routerrpc.Router_HtlcInterceptorClient, error) {
	if len(s.htlcInterceptorMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.htlcInterceptorMockData[0]
	s.htlcInterceptorMockData = s.htlcInterceptorMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewHtlcInterceptorMockData() (<-chan *routerrpc.ForwardHtlcInterceptResponse, chan<- *routerrpc.ForwardHtlcInterceptRequest) {
	sendChan := make(chan *routerrpc.ForwardHtlcInterceptResponse)
	recvChan := make(chan *routerrpc.ForwardHtlcInterceptRequest)
	s.htlcInterceptorMockData = append(s.htlcInterceptorMockData, NewMockHtlcInterceptorClient(sendChan, recvChan))

	return sendChan, recvChan
}

func (s *MockLightningNetworkService) ListChannels(in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	if len(s.listChannelsMockData) == 0 {
		return &lnrpc.ListChannelsResponse{}, errors.New("NotFound")
	}

	response := s.listChannelsMockData[0]
	s.listChannelsMockData = s.listChannelsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetListChannelsMockData(mockData *lnrpc.ListChannelsResponse) {
	s.listChannelsMockData = append(s.listChannelsMockData, mockData)
}

func (s *MockLightningNetworkService) ListPeers(in *lnrpc.ListPeersRequest, opts ...grpc.CallOption) (*lnrpc.ListPeersResponse, error) {
	if len(s.listPeersMockData) == 0 {
		return &lnrpc.ListPeersResponse{}, errors.New("NotFound")
	}

	response := s.listPeersMockData[0]
	s.listPeersMockData = s.listPeersMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetListPeersMockData(mockData *lnrpc.ListPeersResponse) {
	s.listPeersMockData = append(s.listPeersMockData, mockData)
}

func (s *MockLightningNetworkService) OpenChannel(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error) {
	if len(s.openChannelMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.openChannelMockData[0]
	s.openChannelMockData = s.openChannelMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewOpenChannelMockData() chan<- *lnrpc.OpenStatusUpdate {
	recvChan := make(chan *lnrpc.OpenStatusUpdate)
	s.openChannelMockData = append(s.openChannelMockData, NewMockOpenChannelClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) OpenChannelSync(in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error) {
	if len(s.openChannelSyncMockData) == 0 {
		return &lnrpc.ChannelPoint{}, errors.New("NotFound")
	}

	response := s.openChannelSyncMockData[0]
	s.openChannelSyncMockData = s.openChannelSyncMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetOpenChannelSyncMockData(mockData *lnrpc.ChannelPoint) {
	s.openChannelSyncMockData = append(s.openChannelSyncMockData, mockData)
}

func (s *MockLightningNetworkService) PublishTransaction(in *walletrpc.Transaction, opts ...grpc.CallOption) (*walletrpc.PublishResponse, error) {
	if len(s.publishTransactionMockData) == 0 {
		return &walletrpc.PublishResponse{}, errors.New("NotFound")
	}

	response := s.publishTransactionMockData[0]
	s.publishTransactionMockData = s.publishTransactionMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetPublishTransactionMockData(mockData *walletrpc.PublishResponse) {
	s.publishTransactionMockData = append(s.publishTransactionMockData, mockData)
}

func (s *MockLightningNetworkService) RegisterBlockEpochNtfn(in *chainrpc.BlockEpoch, opts ...grpc.CallOption) (chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient, error) {
	if len(s.registerBlockEpochNtfnMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.registerBlockEpochNtfnMockData[0]
	s.registerBlockEpochNtfnMockData = s.registerBlockEpochNtfnMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewRegisterBlockEpochNtfnMockData() chan<- *chainrpc.BlockEpoch {
	recvChan := make(chan *chainrpc.BlockEpoch)
	s.registerBlockEpochNtfnMockData = append(s.registerBlockEpochNtfnMockData, NewMockRegisterBlockEpochNtfnClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SendCustomMessage(in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error) {
	if len(s.sendCustomMessageMockData) == 0 {
		return &lnrpc.SendCustomMessageResponse{}, errors.New("NotFound")
	}

	response := s.sendCustomMessageMockData[0]
	s.sendCustomMessageMockData = s.sendCustomMessageMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetSendCustomMessageMockData(mockData *lnrpc.SendCustomMessageResponse) {
	s.sendCustomMessageMockData = append(s.sendCustomMessageMockData, mockData)
}

func (s *MockLightningNetworkService) SendPaymentV2(in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentV2Client, error) {
	if len(s.sendPaymentV2MockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.sendPaymentV2MockData[0]
	s.sendPaymentV2MockData = s.sendPaymentV2MockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSendPaymentV2MockData() chan<- *lnrpc.Payment {
	recvChan := make(chan *lnrpc.Payment)
	s.sendPaymentV2MockData = append(s.sendPaymentV2MockData, NewMockSendPaymentV2Client(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SignMessage(in *lnrpc.SignMessageRequest, opts ...grpc.CallOption) (*lnrpc.SignMessageResponse, error) {
	if len(s.signMessageMockData) == 0 {
		return &lnrpc.SignMessageResponse{}, errors.New("NotFound")
	}

	response := s.signMessageMockData[0]
	s.signMessageMockData = s.signMessageMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetSignMessageMockData(mockData *lnrpc.SignMessageResponse) {
	s.signMessageMockData = append(s.signMessageMockData, mockData)
}

func (s *MockLightningNetworkService) SubscribeChannelBackups(in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
	if len(s.subscribeChannelBackupsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeChannelBackupsMockData[0]
	s.subscribeChannelBackupsMockData = s.subscribeChannelBackupsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeChannelBackupsMockData() chan<- *lnrpc.ChanBackupSnapshot {
	recvChan := make(chan *lnrpc.ChanBackupSnapshot)
	s.subscribeChannelBackupsMockData = append(s.subscribeChannelBackupsMockData, NewMockSubscribeChannelBackupsClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeChannelEvents(in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	if len(s.subscribeChannelEventsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeChannelEventsMockData[0]
	s.subscribeChannelEventsMockData = s.subscribeChannelEventsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeChannelEventsMockData() chan<- *lnrpc.ChannelEventUpdate {
	recvChan := make(chan *lnrpc.ChannelEventUpdate)
	s.subscribeChannelEventsMockData = append(s.subscribeChannelEventsMockData, NewMockSubscribeChannelEventsClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeChannelGraph(in *lnrpc.GraphTopologySubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {
	if len(s.subscribeChannelGraphMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeChannelGraphMockData[0]
	s.subscribeChannelGraphMockData = s.subscribeChannelGraphMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeChannelGraphMockData() chan<- *lnrpc.GraphTopologyUpdate {
	recvChan := make(chan *lnrpc.GraphTopologyUpdate)
	s.subscribeChannelGraphMockData = append(s.subscribeChannelGraphMockData, NewMockSubscribeChannelGraphClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeCustomMessages(in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	if len(s.subscribeCustomMessagesMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeCustomMessagesMockData[0]
	s.subscribeCustomMessagesMockData = s.subscribeCustomMessagesMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeCustomMessagesMockData() chan<- *lnrpc.CustomMessage {
	recvChan := make(chan *lnrpc.CustomMessage)
	s.subscribeCustomMessagesMockData = append(s.subscribeCustomMessagesMockData, NewMockSubscribeCustomMessagesClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeHtlcEvents(in *routerrpc.SubscribeHtlcEventsRequest, opts ...grpc.CallOption) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	if len(s.subscribeHtlcEventsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeHtlcEventsMockData[0]
	s.subscribeHtlcEventsMockData = s.subscribeHtlcEventsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeHtlcEventsMockData() chan<- *routerrpc.HtlcEvent {
	recvChan := make(chan *routerrpc.HtlcEvent)
	s.subscribeHtlcEventsMockData = append(s.subscribeHtlcEventsMockData, NewMockSubscribeHtlcEventsClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeInvoices(in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	if len(s.subscribeInvoicesMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeInvoicesMockData[0]
	s.subscribeInvoicesMockData = s.subscribeInvoicesMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeInvoicesMockData() chan<- *lnrpc.Invoice {
	recvChan := make(chan *lnrpc.Invoice)
	s.subscribeInvoicesMockData = append(s.subscribeInvoicesMockData, NewMockSubscribeInvoicesClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribePeerEvents(in *lnrpc.PeerEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error) {
	if len(s.subscribePeerEventsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribePeerEventsMockData[0]
	s.subscribePeerEventsMockData = s.subscribePeerEventsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribePeerEventsMockData() chan<- *lnrpc.PeerEvent {
	recvChan := make(chan *lnrpc.PeerEvent)
	s.subscribePeerEventsMockData = append(s.subscribePeerEventsMockData, NewMockSubscribePeerEventsClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) SubscribeTransactions(in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	if len(s.subscribeTransactionsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.subscribeTransactionsMockData[0]
	s.subscribeTransactionsMockData = s.subscribeTransactionsMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) NewSubscribeTransactionsMockData() chan<- *lnrpc.Transaction {
	recvChan := make(chan *lnrpc.Transaction)
	s.subscribeTransactionsMockData = append(s.subscribeTransactionsMockData, NewMockSubscribeTransactionsClient(recvChan))

	return recvChan
}

func (s *MockLightningNetworkService) UpdateChannelPolicy(in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error) {
	if len(s.updateChannelPolicyMockData) == 0 {
		return &lnrpc.PolicyUpdateResponse{}, errors.New("NotFound")
	}

	response := s.updateChannelPolicyMockData[0]
	s.updateChannelPolicyMockData = s.updateChannelPolicyMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetUpdateChannelPolicyMockData(mockData *lnrpc.PolicyUpdateResponse) {
	s.updateChannelPolicyMockData = append(s.updateChannelPolicyMockData, mockData)
}

func (s *MockLightningNetworkService) WalletBalance(in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	if len(s.walletBalanceMockData) == 0 {
		return &lnrpc.WalletBalanceResponse{}, errors.New("NotFound")
	}

	response := s.walletBalanceMockData[0]
	s.walletBalanceMockData = s.walletBalanceMockData[1:]
	return response, nil
}

func (s *MockLightningNetworkService) SetWalletBalanceMockData(mockData *lnrpc.WalletBalanceResponse) {
	s.walletBalanceMockData = append(s.walletBalanceMockData, mockData)
}
