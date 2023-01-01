package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribePeerEventsClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.PeerEvent
}

func NewMockSubscribePeerEventsClient(recvChan <-chan *lnrpc.PeerEvent) lnrpc.Lightning_SubscribePeerEventsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribePeerEventsClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribePeerEventsClient) Recv() (*lnrpc.PeerEvent, error) {
	receive := <-c.recvChan
	return receive, nil
}
