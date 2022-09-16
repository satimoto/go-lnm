package cdr

import (
	"os"

	"github.com/satimoto/go-datastore/pkg/cdr"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/invoicerequest"
	"github.com/satimoto/go-datastore/pkg/promotion"
	"github.com/satimoto/go-lsp/internal/ferp"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/session"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type CdrResolver struct {
	Repository               cdr.CdrRepository
	LightningService         lightningnetwork.LightningNetwork
	NotificationService      notification.Notification
	OcpiService              ocpi.Ocpi
	InvoiceRequestRepository invoicerequest.InvoiceRequestRepository
	PromotionRepository      promotion.PromotionRepository
	SessionResolver          *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService) *CdrResolver {
	ferpService := ferp.NewService(os.Getenv("FERP_RPC_ADDRESS"))

	return NewResolverWithFerpService(repositoryService, ferpService)
}

func NewResolverWithFerpService(repositoryService *db.RepositoryService, ferpService ferp.Ferp) *CdrResolver {
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService(os.Getenv("FCM_API_KEY"))
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, ferpService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, ferpService ferp.Ferp, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *CdrResolver {
	return &CdrResolver{
		Repository:               cdr.NewRepository(repositoryService),
		LightningService:         lightningService,
		OcpiService:              ocpiService,
		NotificationService:      notificationService,
		InvoiceRequestRepository: invoicerequest.NewRepository(repositoryService),
		PromotionRepository:      promotion.NewRepository(repositoryService),
		SessionResolver:          session.NewResolverWithServices(repositoryService, ferpService, lightningService, notificationService, ocpiService),
	}
}
