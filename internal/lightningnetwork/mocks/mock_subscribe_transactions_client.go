package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeTransactionsClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.Transaction
}

func NewMockSubscribeTransactionsClient(recvChan <-chan *lnrpc.Transaction) lnrpc.Lightning_SubscribeTransactionsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeTransactionsClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeTransactionsClient) Recv() (*lnrpc.Transaction, error) {
	receive := <-c.recvChan
	return receive, nil
}

