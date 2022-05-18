package cdr

import (
	"context"
	"os"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/session"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type CdrRepository interface {
	GetCdrByUid(ctx context.Context, uid string) (db.Cdr, error)
}

type CdrResolver struct {
	Repository          CdrRepository
	LightningService    lightningnetwork.LightningNetwork
	NotificationService notification.Notification
	OcpiService         ocpi.Ocpi
	SessionResolver     *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService) *CdrResolver {
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService()
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *CdrResolver {
	repo := CdrRepository(repositoryService)

	return &CdrResolver{
		Repository:          repo,
		LightningService:    lightningService,
		OcpiService:         ocpiService,
		NotificationService: notificationService,
		SessionResolver:     session.NewResolverWithServices(repositoryService, lightningService, notificationService, ocpiService),
	}
}