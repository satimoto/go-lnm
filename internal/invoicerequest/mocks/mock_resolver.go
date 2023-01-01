package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	invoicerequestMocks "github.com/satimoto/go-datastore/pkg/invoicerequest/mocks"
	"github.com/satimoto/go-lsp/internal/invoicerequest"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *invoicerequest.InvoiceRequestResolver {
	return &invoicerequest.InvoiceRequestResolver{
		Repository:          invoicerequestMocks.NewRepository(repositoryService),
		NotificationService: services.NotificationService,
	}
}
