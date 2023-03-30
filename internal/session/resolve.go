package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/location"
	"github.com/satimoto/go-datastore/pkg/session"
	"github.com/satimoto/go-datastore/pkg/token"
	"github.com/satimoto/go-datastore/pkg/tokenauthorization"
	"github.com/satimoto/go-lnm/internal/account"
	"github.com/satimoto/go-lnm/internal/ferp"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	"github.com/satimoto/go-lnm/internal/notification"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/tariff"
	"github.com/satimoto/go-lnm/internal/user"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type SessionResolver struct {
	Repository                   session.SessionRepository
	FerpService                  ferp.Ferp
	LightningService             lightningnetwork.LightningNetwork
	NotificationService          notification.Notification
	OcpiService                  ocpi.Ocpi
	AccountResolver              *account.AccountResolver
	LocationRepository           location.LocationRepository
	TariffResolver               *tariff.TariffResolver
	TokenRepository              token.TokenRepository
	TokenAuthorizationRepository tokenauthorization.TokenAuthorizationRepository
	UserResolver                 *user.UserResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *SessionResolver {
	return &SessionResolver{
		Repository:                   session.NewRepository(repositoryService),
		FerpService:                  services.FerpService,
		LightningService:             services.LightningService,
		OcpiService:                  services.OcpiService,
		NotificationService:          services.NotificationService,
		AccountResolver:              account.NewResolver(repositoryService),
		LocationRepository:           location.NewRepository(repositoryService),
		TariffResolver:               tariff.NewResolver(repositoryService),
		TokenRepository:              token.NewRepository(repositoryService),
		TokenAuthorizationRepository: tokenauthorization.NewRepository(repositoryService),
		UserResolver:                 user.NewResolver(repositoryService, services),
	}
}
