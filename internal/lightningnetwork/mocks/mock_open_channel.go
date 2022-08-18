package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockOpenChannelClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.OpenStatusUpdate
}

func NewMockOpenChannelClient(recvChan <-chan *lnrpc.OpenStatusUpdate) lnrpc.Lightning_OpenChannelClient {
	clientStream := NewMockClientStream()
	return &MockOpenChannelClient{
		ClientStream: clientStream,
		recvChan:     recvChan,
	}
}

func (c *MockOpenChannelClient) Recv() (*lnrpc.OpenStatusUpdate, error) {
	receive := <-c.recvChan
	return receive, nil
}
