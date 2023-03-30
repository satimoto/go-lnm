package mocks

import (
	ferp "github.com/satimoto/go-lnm/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lnm/internal/lightningnetwork/mocks"
	notification "github.com/satimoto/go-lnm/internal/notification/mocks"
	"github.com/satimoto/go-lnm/internal/service"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewService(ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *service.ServiceResolver {
	return &service.ServiceResolver{
		FerpService:         ferpService,
		LightningService:    lightningService,
		NotificationService: notificationService,
		OcpiService:         ocpiService,
	}
}
