package mocks

import (
	"github.com/satimoto/go-lsp/lsprpc"
)

type MockLspService struct {
	openChannelMockData []*lsprpc.OpenChannelResponse
	listChannelsMockData []*lsprpc.ListChannelsResponse
}

func NewService() *MockLspService {
	return &MockLspService{}
}
