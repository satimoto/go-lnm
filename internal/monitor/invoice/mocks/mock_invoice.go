package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
)

func NewInvoiceMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *invoice.InvoiceMonitor {
	return &invoice.InvoiceMonitor{
		LightningService: lightningService,
		SessionResolver:  session.NewResolver(repositoryService, lightningService, notificationService, ocpiService),
	}
}
