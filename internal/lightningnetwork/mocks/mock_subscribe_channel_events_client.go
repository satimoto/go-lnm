package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeChannelEventsClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.ChannelEventUpdate
}

func NewMockSubscribeChannelEventsClient(recvChan <-chan *lnrpc.ChannelEventUpdate) lnrpc.Lightning_SubscribeChannelEventsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeChannelEventsClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeChannelEventsClient) Recv() (*lnrpc.ChannelEventUpdate, error) {
	receive := <-c.recvChan
	return receive, nil
}
