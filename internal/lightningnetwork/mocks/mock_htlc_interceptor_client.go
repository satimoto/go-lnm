package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockHtlcInterceptorClient struct {
	grpc.ClientStream
}

func NewMockHtlcInterceptorClient() routerrpc.Router_HtlcInterceptorClient {
	clientStream := NewMockClientStream()
	return &MockHtlcInterceptorClient{
		ClientStream: clientStream,
	}
}

func (c *MockHtlcInterceptorClient) Send(*routerrpc.ForwardHtlcInterceptResponse) error {
	return nil
}

func (c *MockHtlcInterceptorClient) Recv() (*routerrpc.ForwardHtlcInterceptRequest, error) {
	return &routerrpc.ForwardHtlcInterceptRequest{}, nil
}

