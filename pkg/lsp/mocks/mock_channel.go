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
