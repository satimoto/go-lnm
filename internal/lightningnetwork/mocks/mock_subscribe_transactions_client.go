package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeTransactionsClient struct {
	grpc.ClientStream
}

func NewMockSubscribeTransactionsClient() lnrpc.Lightning_SubscribeTransactionsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeTransactionsClient{
		ClientStream: clientStream,
	}
}

func (c *MockSubscribeTransactionsClient) Recv() (*lnrpc.Transaction, error) {
	return &lnrpc.Transaction{}, nil
}

