package invoice

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/invoicerequest"
	"github.com/satimoto/go-datastore/pkg/session"
	"github.com/satimoto/go-datastore/pkg/tokenauthorization"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	"github.com/satimoto/go-lnm/internal/service"
)

type RpcInvoiceResolver struct {
	LightningService             lightningnetwork.LightningNetwork
	InvoiceRequestRepository     invoicerequest.InvoiceRequestRepository
	SessionRepository            session.SessionRepository
	TokenAuthorizationRepository tokenauthorization.TokenAuthorizationRepository
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcInvoiceResolver {
	return &RpcInvoiceResolver{
		LightningService:             services.LightningService,
		InvoiceRequestRepository:     invoicerequest.NewRepository(repositoryService),
		SessionRepository:            session.NewRepository(repositoryService),
		TokenAuthorizationRepository: tokenauthorization.NewRepository(repositoryService),
	}
}
