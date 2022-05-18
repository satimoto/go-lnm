package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
)

func NewHtlcMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, customMessageMonitor *custommessage.CustomMessageMonitor) *htlc.HtlcMonitor {
	return &htlc.HtlcMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		CustomMessageMonitor:   customMessageMonitor,
	}
}
