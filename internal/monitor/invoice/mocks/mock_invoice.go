package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	ocpi "github.com/satimoto/go-lsp/internal/ocpi/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
)

func NewInvoiceMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, ocpiService *ocpi.MockOcpiService) *invoice.InvoiceMonitor {
	return &invoice.InvoiceMonitor{
		LightningService: lightningService,
		SessionResolver:  session.NewResolver(repositoryService, lightningService, ocpiService),
	}
}
