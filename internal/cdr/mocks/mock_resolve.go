package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/internal/cdr"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, ocpiService *ocpi.MockOcpiService) *cdr.CdrResolver {
	notificationService := notification.NewService()
	repo := cdr.CdrRepository(repositoryService)

	return &cdr.CdrResolver{
		Repository:          repo,
		LightningService:    lightningService,
		NotificationService: notificationService,
		OcpiService:         ocpiService,
		SessionResolver:     session.NewResolver(repositoryService, lightningService, notificationService, ocpiService),
	}
}
