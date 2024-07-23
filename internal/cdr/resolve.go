package cdr

import (
	"github.com/satimoto/go-datastore/pkg/cdr"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/invoicerequest"
	"github.com/satimoto/go-datastore/pkg/pendingnotification"
	"github.com/satimoto/go-datastore/pkg/promotion"
	"github.com/satimoto/go-lnm/internal/ferp"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	"github.com/satimoto/go-lnm/internal/notification"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/session"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type CdrResolver struct {
	Repository                    cdr.CdrRepository
	FerpService                   ferp.Ferp
	LightningService              lightningnetwork.LightningNetwork
	NotificationService           notification.Notification
	OcpiService                   ocpi.Ocpi
	InvoiceRequestRepository      invoicerequest.InvoiceRequestRepository
	PendingNotificationRepository pendingnotification.PendingNotificationRepository
	PromotionRepository           promotion.PromotionRepository
	SessionResolver               *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *CdrResolver {
	return &CdrResolver{
		Repository:                    cdr.NewRepository(repositoryService),
		FerpService:                   services.FerpService,
		LightningService:              services.LightningService,
		NotificationService:           services.NotificationService,
		OcpiService:                   services.OcpiService,
		InvoiceRequestRepository:      invoicerequest.NewRepository(repositoryService),
		PendingNotificationRepository: pendingnotification.NewRepository(repositoryService),
		PromotionRepository:           promotion.NewRepository(repositoryService),
		SessionResolver:               session.NewResolver(repositoryService, services),
	}
}
