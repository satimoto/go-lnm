package cdr

import (
	"os"

	"github.com/satimoto/go-datastore/pkg/cdr"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/exchange"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/session"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type CdrResolver struct {
	Repository          cdr.CdrRepository
	LightningService    lightningnetwork.LightningNetwork
	NotificationService notification.Notification
	OcpiService         ocpi.Ocpi
	SessionResolver     *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, exchangeService exchange.Exchange) *CdrResolver {
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService(os.Getenv("FCM_API_KEY"))
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, exchangeService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, exchangeService exchange.Exchange, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *CdrResolver {
	return &CdrResolver{
		Repository:          cdr.NewRepository(repositoryService),
		LightningService:    lightningService,
		OcpiService:         ocpiService,
		NotificationService: notificationService,
		SessionResolver:     session.NewResolverWithServices(repositoryService, exchangeService, lightningService, notificationService, ocpiService),
	}
}
