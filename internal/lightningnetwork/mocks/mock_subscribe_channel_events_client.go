package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeChannelEventsClient struct {
	grpc.ClientStream
}

func NewMockSubscribeChannelEventsClient() lnrpc.Lightning_SubscribeChannelEventsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeChannelEventsClient{
		ClientStream: clientStream,
	}
}

func (c *MockSubscribeChannelEventsClient) Recv() (*lnrpc.ChannelEventUpdate, error) {
	return &lnrpc.ChannelEventUpdate{}, nil
}
