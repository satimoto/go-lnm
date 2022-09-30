package mocks

import (
	"context"
	"errors"
	"sync"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/service"
)

type OpenChannelMockData struct {
	OpenChannelRequest *lnrpc.OpenChannelRequest
	ChannelRequest     db.ChannelRequest
}

type MockPsbtFundService struct {
	openChannelRequestMockData  []OpenChannelMockData
	openChannelResponseMockData []error
}

func NewService(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *MockPsbtFundService {
	return &MockPsbtFundService{}
}

func (s *MockPsbtFundService) Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
}

func (s *MockPsbtFundService) OpenChannel(ctx context.Context, request *lnrpc.OpenChannelRequest, channelRequest db.ChannelRequest) error {
	s.openChannelRequestMockData = append(s.openChannelRequestMockData, OpenChannelMockData{
		OpenChannelRequest: request,
		ChannelRequest:     channelRequest,
	})

	if len(s.openChannelResponseMockData) == 0 {
		return nil
	}

	response := s.openChannelResponseMockData[0]
	s.openChannelResponseMockData = s.openChannelResponseMockData[1:]
	return response
}

func (s *MockPsbtFundService) GetOpenChannelMockData() (OpenChannelMockData, error) {
	if len(s.openChannelRequestMockData) == 0 {
		return OpenChannelMockData{}, errors.New("NotFound")
	}

	response := s.openChannelRequestMockData[0]
	s.openChannelRequestMockData = s.openChannelRequestMockData[1:]
	return response, nil
}

func (s *MockPsbtFundService) SetOpenChannelMockData(err error) {
	s.openChannelResponseMockData = append(s.openChannelResponseMockData, err)
}
