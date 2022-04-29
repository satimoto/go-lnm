package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
)

func NewChannelEventMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *channelevent.ChannelEventMonitor {
	return &channelevent.ChannelEventMonitor{
		LightningService: lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}