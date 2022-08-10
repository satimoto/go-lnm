package mocks

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockClientStream struct{}

func NewMockClientStream() grpc.ClientStream {
	return &MockClientStream{}
}

func (s *MockClientStream) Header() (metadata.MD, error) {
	return metadata.MD{}, nil
}

func (s *MockClientStream) Trailer() metadata.MD {
	return metadata.MD{}
}

func (s *MockClientStream) CloseSend() error {
	return nil
}

func (s *MockClientStream) Context() context.Context {
	return context.Background()
}

func (s *MockClientStream) SendMsg(m interface{}) error {
	return nil
}

func (s *MockClientStream) RecvMsg(m interface{}) error {
	return nil
}
