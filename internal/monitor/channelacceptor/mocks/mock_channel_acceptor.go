package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelacceptor"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewChannelAcceptorMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *channelacceptor.ChannelAcceptorMonitor {
	return &channelacceptor.ChannelAcceptorMonitor{
		LightningService: services.LightningService,
	}
}
