package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	ferp "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewInvoiceMonitor(repositoryService *mocks.MockRepositoryService, ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *invoice.InvoiceMonitor {
	return &invoice.InvoiceMonitor{
		LightningService: lightningService,
		SessionResolver:  session.NewResolver(repositoryService, ferpService, lightningService, notificationService, ocpiService),
	}
}
