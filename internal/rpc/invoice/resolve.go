package invoice

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/invoicerequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
)

type RpcInvoiceResolver struct {
	LightningService       lightningnetwork.LightningNetwork
	InvoiceRequestResolver *invoicerequest.InvoiceRequestResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcInvoiceResolver {

	return &RpcInvoiceResolver{
		LightningService:       services.LightningService,
		InvoiceRequestResolver: invoicerequest.NewResolver(repositoryService, services),
	}
}
