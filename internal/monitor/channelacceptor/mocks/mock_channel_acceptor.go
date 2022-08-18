package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelacceptor"
)

func NewChannelAcceptorMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *channelacceptor.ChannelAcceptorMonitor {
	return &channelacceptor.ChannelAcceptorMonitor{
		LightningService: lightningService,
	}
}
