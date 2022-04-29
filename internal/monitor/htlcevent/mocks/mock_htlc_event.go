package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
)

func NewHtlcEventMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *htlcevent.HtlcEventMonitor {
	return &htlcevent.HtlcEventMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
