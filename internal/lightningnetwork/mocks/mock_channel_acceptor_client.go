package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockChannelAcceptorClient struct {
	grpc.ClientStream
	sendChan chan<- *lnrpc.ChannelAcceptResponse
	recvChan <-chan *lnrpc.ChannelAcceptRequest
}

func NewMockChannelAcceptorClient(sendChan chan<- *lnrpc.ChannelAcceptResponse, recvChan <-chan *lnrpc.ChannelAcceptRequest) lnrpc.Lightning_ChannelAcceptorClient {
	clientStream := NewMockClientStream()
	return &MockChannelAcceptorClient{
		ClientStream: clientStream,
		sendChan:     sendChan,
		recvChan:     recvChan,
	}
}

func (c *MockChannelAcceptorClient) Send(send *lnrpc.ChannelAcceptResponse) error {
	c.sendChan <- send
	return nil
}

func (c *MockChannelAcceptorClient) Recv() (*lnrpc.ChannelAcceptRequest, error) {
	receive := <-c.recvChan
	return receive, nil
}
