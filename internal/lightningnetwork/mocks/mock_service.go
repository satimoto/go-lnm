package mocks

import (
	"encoding/hex"
	"errors"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockLightningNetworkService struct {
	addInvoiceMockData              []*lnrpc.Invoice
	getInfoMockData                 []*lnrpc.GetInfoResponse
	htlcInterceptorMockData         []routerrpc.Router_HtlcInterceptorClient
	openChannelSyncMockData         []*lnrpc.ChannelPoint
	sendCustomMessageMockData       []*lnrpc.SendCustomMessageResponse
	subscribeChannelBackupsMockData []lnrpc.Lightning_SubscribeChannelBackupsClient
	subscribeChannelEventsMockData  []lnrpc.Lightning_SubscribeChannelEventsClient
	subscribeCustomMessagesMockData []lnrpc.Lightning_SubscribeCustomMessagesClient
	subscribeHtlcEventsMockData     []routerrpc.Router_SubscribeHtlcEventsClient
	subscribeInvoicesMockData       []lnrpc.Lightning_SubscribeInvoicesClient
	subscribeTransactionsMockData   []lnrpc.Lightning_SubscribeTransactionsClient
}

func NewService() *MockLightningNetworkService {
	return &MockLightningNetworkService{}
}

func (s *MockLightningNetworkService) AddInvoice(in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	s.addInvoiceMockData = append(s.addInvoiceMockData, in)

	return &lnrpc.AddInvoiceResponse{
		RHash:          in.RHash,
		PaymentRequest: hex.EncodeToString(in.RPreimage),
	}, nil
}

func (s *MockLightningNetworkService) GetAddInvoiceMockData() (*lnrpc.Invoice, error) {
	if len(s.addInvoiceMockData) == 0 {
		return &lnrpc.Invoice{}, errors.New("NotFound")
	}

	response := s.addInvoiceMockData[0]
	s.addInvoiceMockData = s.addInvoiceMockData[1:]
	return response, nil
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
