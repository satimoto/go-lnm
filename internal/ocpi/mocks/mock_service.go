package mocks

import (
	"context"
	"errors"

	"github.com/satimoto/go-ocpi-api/ocpirpc/commandrpc"
	"github.com/satimoto/go-ocpi-api/ocpirpc/tokenrpc"
	"google.golang.org/grpc"
)

type MockOcpiService struct {
	stopSessionMockData  []*commandrpc.StopSessionResponse
	updateTokensMockData []*tokenrpc.UpdateTokensResponse
}

func NewService() *MockOcpiService {
	return &MockOcpiService{}
}

func (s *MockOcpiService) StopSession(ctx context.Context, in *commandrpc.StopSessionRequest, opts ...grpc.CallOption) (*commandrpc.StopSessionResponse, error) {
	if len(s.stopSessionMockData) == 0 {
		return &commandrpc.StopSessionResponse{}, errors.New("NotFound")
	}

	response := s.stopSessionMockData[0]
	s.stopSessionMockData = s.stopSessionMockData[1:]
	return response, nil
}

func (s *MockOcpiService) SetStopSessionMockData(mockData *commandrpc.StopSessionResponse) {
	s.stopSessionMockData = append(s.stopSessionMockData, mockData)
}

func (s *MockOcpiService) UpdateTokens(ctx context.Context, in *tokenrpc.UpdateTokensRequest, opts ...grpc.CallOption) (*tokenrpc.UpdateTokensResponse, error) {
	if len(s.updateTokensMockData) == 0 {
		return &tokenrpc.UpdateTokensResponse{}, errors.New("NotFound")
	}

	response := s.updateTokensMockData[0]
	s.updateTokensMockData = s.updateTokensMockData[1:]
	return response, nil
}

func (s *MockOcpiService) SetUpdateTokensMockData(mockData *tokenrpc.UpdateTokensResponse) {
	s.updateTokensMockData = append(s.updateTokensMockData, mockData)
}