package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	"github.com/satimoto/go-lsp/internal/service"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
)

func NewInvoiceMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *invoice.InvoiceMonitor {
	return &invoice.InvoiceMonitor{
		LightningService: services.LightningService,
		SessionResolver:  session.NewResolver(repositoryService, services),
	}
}
