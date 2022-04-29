package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockSubscribeHtlcEventsClient struct {
	grpc.ClientStream
	recvChan <-chan *routerrpc.HtlcEvent
}

func NewMockSubscribeHtlcEventsClient(recvChan <-chan *routerrpc.HtlcEvent) routerrpc.Router_SubscribeHtlcEventsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeHtlcEventsClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeHtlcEventsClient) Recv() (*routerrpc.HtlcEvent, error) {
	receive := <-c.recvChan
	return receive, nil
}

