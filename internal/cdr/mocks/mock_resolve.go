package mocks

import (
	cdrMocks "github.com/satimoto/go-datastore/pkg/cdr/mocks"
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	invoicerequest "github.com/satimoto/go-datastore/pkg/invoicerequest/mocks"
	promotion "github.com/satimoto/go-datastore/pkg/promotion/mocks"
	"github.com/satimoto/go-lsp/internal/cdr"
	ferp "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, ocpiService *ocpi.MockOcpiService) *cdr.CdrResolver {
	notificationService := notification.NewService()

	return &cdr.CdrResolver{
		Repository:               cdrMocks.NewRepository(repositoryService),
		FerpService:              ferpService,
		LightningService:         lightningService,
		NotificationService:      notificationService,
		OcpiService:              ocpiService,
		InvoiceRequestRepository: invoicerequest.NewRepository(repositoryService),
		PromotionRepository:      promotion.NewRepository(repositoryService),
		SessionResolver:          session.NewResolver(repositoryService, ferpService, lightningService, notificationService, ocpiService),
	}
}
