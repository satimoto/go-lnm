package session

import (
	"os"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/location"
	"github.com/satimoto/go-datastore/pkg/session"
	"github.com/satimoto/go-lsp/internal/countryaccount"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type SessionResolver struct {
	Repository             session.SessionRepository
	LightningService       lightningnetwork.LightningNetwork
	NotificationService    notification.Notification
	OcpiService            ocpi.Ocpi
	CountryAccountResolver *countryaccount.CountryAccountResolver
	LocationRepository     location.LocationRepository
	TariffResolver         *tariff.TariffResolver
	UserResolver           *user.UserResolver
}

func NewResolver(repositoryService *db.RepositoryService) *SessionResolver {
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService()
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *SessionResolver {
	return &SessionResolver{
		Repository:             session.NewRepository(repositoryService),
		LightningService:       lightningService,
		OcpiService:            ocpiService,
		NotificationService:    notificationService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationRepository:     location.NewRepository(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolverWithServices(repositoryService, ocpiService),
	}
}
