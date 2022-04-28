package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockSubscribeHtlcEventsClient struct {
	grpc.ClientStream
}

func NewMockSubscribeHtlcEventsClient() routerrpc.Router_SubscribeHtlcEventsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeHtlcEventsClient{
		ClientStream: clientStream,
	}
}

func (c *MockSubscribeHtlcEventsClient) Recv() (*routerrpc.HtlcEvent, error) {
	return &routerrpc.HtlcEvent{}, nil
}

