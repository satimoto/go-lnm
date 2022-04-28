package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeInvoicesClient struct {
	grpc.ClientStream
}

func NewMockSubscribeInvoicesClient() lnrpc.Lightning_SubscribeInvoicesClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeInvoicesClient{
		ClientStream: clientStream,
	}
}

func (c *MockSubscribeInvoicesClient) Recv() (*lnrpc.Invoice, error) {
	return &lnrpc.Invoice{}, nil
}

