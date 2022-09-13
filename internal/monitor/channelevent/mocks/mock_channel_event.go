package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewChannelEventMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, htlcMonitor *htlc.HtlcMonitor) *channelevent.ChannelEventMonitor {
	return &channelevent.ChannelEventMonitor{
		LightningService:       lightningService,
		HtlcMonitor:            htlcMonitor,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeRepository:         node.NewRepository(repositoryService),
		UserResolver:           user.NewResolver(repositoryService),
	}
}
