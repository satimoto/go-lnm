package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	ferp "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
)

func NewHtlcEventMonitor(repositoryService *mocks.MockRepositoryService, ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService) *htlcevent.HtlcEventMonitor {
	return &htlcevent.HtlcEventMonitor{
		FerpService:            ferpService,
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
