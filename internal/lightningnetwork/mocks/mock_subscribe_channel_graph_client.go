package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeChannelGraphClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.GraphTopologyUpdate
}

func NewMockSubscribeChannelGraphClient(recvChan <-chan *lnrpc.GraphTopologyUpdate) lnrpc.Lightning_SubscribeChannelGraphClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeChannelGraphClient{
		ClientStream: clientStream,
		recvChan:     recvChan,
	}
}

func (c *MockSubscribeChannelGraphClient) Recv() (*lnrpc.GraphTopologyUpdate, error) {
	receive := <-c.recvChan
	return receive, nil
}
