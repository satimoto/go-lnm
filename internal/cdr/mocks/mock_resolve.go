package mocks

import (
	cdrMocks "github.com/satimoto/go-datastore/pkg/cdr/mocks"
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	invoicerequest "github.com/satimoto/go-datastore/pkg/invoicerequest/mocks"
	promotion "github.com/satimoto/go-datastore/pkg/promotion/mocks"
	"github.com/satimoto/go-lnm/internal/cdr"
	"github.com/satimoto/go-lnm/internal/service"
	session "github.com/satimoto/go-lnm/internal/session/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *cdr.CdrResolver {
	return &cdr.CdrResolver{
		Repository:               cdrMocks.NewRepository(repositoryService),
		FerpService:              services.FerpService,
		LightningService:         services.LightningService,
		NotificationService:      services.NotificationService,
		OcpiService:              services.OcpiService,
		InvoiceRequestRepository: invoicerequest.NewRepository(repositoryService),
		PromotionRepository:      promotion.NewRepository(repositoryService),
		SessionResolver:          session.NewResolver(repositoryService, services),
	}
}
