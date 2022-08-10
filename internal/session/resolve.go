package session

import (
	"os"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/location"
	"github.com/satimoto/go-datastore/pkg/session"
	"github.com/satimoto/go-lsp/internal/countryaccount"
	"github.com/satimoto/go-lsp/internal/ferp"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type SessionResolver struct {
	Repository             session.SessionRepository
	FerpService            ferp.Ferp
	LightningService       lightningnetwork.LightningNetwork
	NotificationService    notification.Notification
	OcpiService            ocpi.Ocpi
	CountryAccountResolver *countryaccount.CountryAccountResolver
	LocationRepository     location.LocationRepository
	TariffResolver         *tariff.TariffResolver
	UserResolver           *user.UserResolver
}

func NewResolver(repositoryService *db.RepositoryService) *SessionResolver {
	ferpService := ferp.NewService(os.Getenv("FERP_RPC_ADDRESS"))

	return NewResolverWithFerpService(repositoryService, ferpService)
}

func NewResolverWithFerpService(repositoryService *db.RepositoryService, ferpService ferp.Ferp) *SessionResolver {
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService(os.Getenv("FCM_API_KEY"))
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, ferpService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, ferpService ferp.Ferp, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *SessionResolver {
	return &SessionResolver{
		Repository:             session.NewRepository(repositoryService),
		FerpService:            ferpService,
		LightningService:       lightningService,
		OcpiService:            ocpiService,
		NotificationService:    notificationService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationRepository:     location.NewRepository(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolverWithServices(repositoryService, ocpiService),
	}
}
