package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"google.golang.org/grpc"
)

type MockRegisterBlockEpochNtfnClient struct {
	grpc.ClientStream
	recvChan <-chan *chainrpc.BlockEpoch
}

func NewMockRegisterBlockEpochNtfnClient(recvChan <-chan *chainrpc.BlockEpoch) chainrpc.ChainNotifier_RegisterBlockEpochNtfnClient {
	clientStream := NewMockClientStream()
	return &MockRegisterBlockEpochNtfnClient{
		ClientStream: clientStream,
		recvChan:     recvChan,
	}
}

func (c *MockRegisterBlockEpochNtfnClient) Recv() (*chainrpc.BlockEpoch, error) {
	receive := <-c.recvChan
	return receive, nil
}
