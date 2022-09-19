package invoicerequest

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/invoicerequest"
)

type InvoiceRequestResolver struct {
	Repository invoicerequest.InvoiceRequestRepository
}

func NewResolver(repositoryService *db.RepositoryService) *InvoiceRequestResolver {
	return &InvoiceRequestResolver{
		Repository: invoicerequest.NewRepository(repositoryService),
	}
}
