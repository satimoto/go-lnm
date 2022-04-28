package mocks

import (
	"context"

	"github.com/satimoto/go-lsp/internal/ocpi"
	"github.com/satimoto/go-ocpi-api/ocpirpc/commandrpc"
	"github.com/satimoto/go-ocpi-api/ocpirpc/tokenrpc"
	"google.golang.org/grpc"
)

type MockOcpiService struct{}

func NewService() ocpi.Ocpi {
	return &MockOcpiService{}
}

func (s *MockOcpiService) StopSession(ctx context.Context, in *commandrpc.StopSessionRequest, opts ...grpc.CallOption) (*commandrpc.StopSessionResponse, error) {
	return &commandrpc.StopSessionResponse{}, nil
}

func (s *MockOcpiService) UpdateTokens(ctx context.Context, in *tokenrpc.UpdateTokensRequest, opts ...grpc.CallOption) (*tokenrpc.UpdateTokensResponse, error) {
	return &tokenrpc.UpdateTokensResponse{}, nil
}
