package mocks

import (
	channelrequestMocks "github.com/satimoto/go-datastore/pkg/channelrequest/mocks"
	"github.com/satimoto/go-datastore/pkg/db/mocks"
	pendingnotificationMocks "github.com/satimoto/go-datastore/pkg/pendingnotification/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/pendingnotification"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewPendingNotificationMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *pendingnotification.PendingNotificationMonitor {
	return &pendingnotification.PendingNotificationMonitor{
		LightningService:              services.LightningService,
		NotificationService:           services.NotificationService,
		ChannelRequestRepository:      channelrequestMocks.NewRepository(repositoryService),
		PendingNotificationRepository: pendingnotificationMocks.NewRepository(repositoryService),
	}
}
