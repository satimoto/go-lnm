package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/service"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewChannelEventMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver, htlcMonitor *htlc.HtlcMonitor) *channelevent.ChannelEventMonitor {
	return &channelevent.ChannelEventMonitor{
		LightningService:       services.LightningService,
		HtlcMonitor:            htlcMonitor,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeRepository:         node.NewRepository(repositoryService),
		UserResolver:           user.NewResolver(repositoryService, services),
	}
}
