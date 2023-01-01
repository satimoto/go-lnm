package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type MockSendPaymentV2Client struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.Payment
}

func NewMockSendPaymentV2Client(recvChan <-chan *lnrpc.Payment) routerrpc.Router_SendPaymentV2Client {
	clientStream := NewMockClientStream()
	return &MockSendPaymentV2Client{
		ClientStream: clientStream,
		recvChan:     recvChan,
	}
}

func (c *MockSendPaymentV2Client) Recv() (*lnrpc.Payment, error) {
	receive := <-c.recvChan
	return receive, nil
}
