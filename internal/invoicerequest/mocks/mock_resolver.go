package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	invoicerequestMocks "github.com/satimoto/go-datastore/pkg/invoicerequest/mocks"
	"github.com/satimoto/go-lsp/internal/invoicerequest"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *invoicerequest.InvoiceRequestResolver {
	return &invoicerequest.InvoiceRequestResolver{
		Repository: invoicerequestMocks.NewRepository(repositoryService),
	}
}
