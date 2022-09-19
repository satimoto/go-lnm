package invoice

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/invoicerequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
)

type RpcInvoiceResolver struct {
	LightningService       lightningnetwork.LightningNetwork
	InvoiceRequestResolver *invoicerequest.InvoiceRequestResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcInvoiceResolver {
	lightningService := lightningnetwork.NewService()

	return NewResolverWithServices(repositoryService, lightningService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *RpcInvoiceResolver {
	return &RpcInvoiceResolver{
		LightningService:       lightningService,
		InvoiceRequestResolver: invoicerequest.NewResolver(repositoryService),
	}
}
