package mocks

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

type MockSubscribeChannelBackupsClient struct {
	grpc.ClientStream
	recvChan <-chan *lnrpc.ChanBackupSnapshot
}

func NewMockSubscribeChannelBackupsClient(recvChan <-chan *lnrpc.ChanBackupSnapshot) lnrpc.Lightning_SubscribeChannelBackupsClient {
	clientStream := NewMockClientStream()
	return &MockSubscribeChannelBackupsClient{
		ClientStream: clientStream,
		recvChan: recvChan,
	}
}

func (c *MockSubscribeChannelBackupsClient) Recv() (*lnrpc.ChanBackupSnapshot, error) {
	receive := <-c.recvChan
	return receive, nil
}
