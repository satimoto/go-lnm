package mocks

import (
	"github.com/satimoto/go-lnm/lsprpc"
)

type MockLspService struct {
	openChannelMockData  []*lsprpc.OpenChannelResponse
	listChannelsMockData []*lsprpc.ListChannelsResponse
}

func NewService() *MockLspService {
	return &MockLspService{}
}
