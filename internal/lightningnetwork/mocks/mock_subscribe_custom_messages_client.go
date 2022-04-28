package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeCustomMessagesClient struct {
	grpc.ClientStream
}

func NewMockSubscribeCustomMessagesClient() lnrpc.Lightning_SubscribeCustomMessagesClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeCustomMessagesClient{
		ClientStream: clientStream,
	}
}

func (c *MockSubscribeCustomMessagesClient) Recv() (*lnrpc.CustomMessage, error) {
	return &lnrpc.CustomMessage{}, nil
}

