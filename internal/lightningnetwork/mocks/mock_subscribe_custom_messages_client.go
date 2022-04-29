package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeCustomMessagesClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.CustomMessage
}

func NewMockSubscribeCustomMessagesClient(recvChan <-chan *lnrpc.CustomMessage) lnrpc.Lightning_SubscribeCustomMessagesClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeCustomMessagesClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeCustomMessagesClient) Recv() (*lnrpc.CustomMessage, error) {
	receive := <-c.recvChan
	return receive, nil
}

