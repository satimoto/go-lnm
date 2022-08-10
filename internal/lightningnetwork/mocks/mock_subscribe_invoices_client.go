package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeInvoicesClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.Invoice
}

func NewMockSubscribeInvoicesClient(recvChan <-chan *lnrpc.Invoice) lnrpc.Lightning_SubscribeInvoicesClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeInvoicesClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeInvoicesClient) Recv() (*lnrpc.Invoice, error) {
	receive := <-c.recvChan
	return receive, nil
}

