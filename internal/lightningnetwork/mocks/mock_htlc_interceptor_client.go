package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockHtlcInterceptorClient struct {
	grpc.ClientStream
	sendChan chan<- *routerrpc.ForwardHtlcInterceptResponse
	recvChan <-chan *routerrpc.ForwardHtlcInterceptRequest
}

func NewMockHtlcInterceptorClient(sendChan chan<- *routerrpc.ForwardHtlcInterceptResponse, recvChan <-chan *routerrpc.ForwardHtlcInterceptRequest) routerrpc.Router_HtlcInterceptorClient {
	clientStream := NewMockClientStream()
	return &MockHtlcInterceptorClient{
		ClientStream: clientStream,
		sendChan: sendChan,
		recvChan: recvChan,
	}
}

func (c *MockHtlcInterceptorClient) Send(send *routerrpc.ForwardHtlcInterceptResponse) error {
	c.sendChan<- send
	return nil
}

func (c *MockHtlcInterceptorClient) Recv() (*routerrpc.ForwardHtlcInterceptRequest, error) {
	receive := <-c.recvChan
	return receive, nil
}

