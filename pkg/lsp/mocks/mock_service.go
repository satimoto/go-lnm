package mocks

import (
	"github.com/satimoto/go-lsp/lsprpc"
)

type MockLspService struct {
	openChannelMockData []*lsprpc.OpenChannelResponse
}

func NewService() *MockLspService {
	return &MockLspService{}
}
