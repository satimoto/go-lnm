package mocks

import (
	"context"
	"errors"

	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
)

func (s *MockLspService) OpenChannel(ctx context.Context, in *lsprpc.OpenChannelRequest, opts ...grpc.CallOption) (*lsprpc.OpenChannelResponse, error) {
	if len(s.openChannelMockData) == 0 {
		return &lsprpc.OpenChannelResponse{}, errors.New("NotFound")
	}

	response := s.openChannelMockData[0]
	s.openChannelMockData = s.openChannelMockData[1:]
	return response, nil
}

func (s *MockLspService) SetOpenChannelMockData(mockData *lsprpc.OpenChannelResponse) {
	s.openChannelMockData = append(s.openChannelMockData, mockData)
}

func (s *MockLspService) ListChannels(ctx context.Context, in *lsprpc.ListChannelsRequest, opts ...grpc.CallOption) (*lsprpc.ListChannelsResponse, error) {
	if len(s.listChannelsMockData) == 0 {
		return &lsprpc.ListChannelsResponse{}, errors.New("NotFound")
	}

	response := s.listChannelsMockData[0]
	s.listChannelsMockData = s.listChannelsMockData[1:]
	return response, nil
}

func (s *MockLspService) SetListChannelsMockData(mockData *lsprpc.ListChannelsResponse) {
	s.listChannelsMockData = append(s.listChannelsMockData, mockData)
}
