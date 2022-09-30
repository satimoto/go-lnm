package mocks

import (
	ferp "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	"github.com/satimoto/go-lsp/internal/service"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewService(ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *service.ServiceResolver {
	return &service.ServiceResolver{
		FerpService:                  ferpService,
		LightningService:             lightningService,
		NotificationService:          notificationService,
		OcpiService:                  ocpiService,
	}
}
