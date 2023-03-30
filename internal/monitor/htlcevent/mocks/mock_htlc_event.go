package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	routingevent "github.com/satimoto/go-datastore/pkg/routingevent/mocks"
	"github.com/satimoto/go-lnm/internal/monitor/htlcevent"
	"github.com/satimoto/go-lnm/internal/service"
)

func NewHtlcEventMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *htlcevent.HtlcEventMonitor {
	return &htlcevent.HtlcEventMonitor{
		FerpService:            services.FerpService,
		LightningService:       services.LightningService,
		RoutingEventRepository: routingevent.NewRepository(repositoryService),
	}
}
