package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	routingevent "github.com/satimoto/go-datastore/pkg/routingevent/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewHtlcEventMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *htlcevent.HtlcEventMonitor {
	return &htlcevent.HtlcEventMonitor{
		FerpService:            services.FerpService,
		LightningService:       services.LightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		RoutingEventRepository: routingevent.NewRepository(repositoryService),
	}
}
