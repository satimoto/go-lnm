package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	location "github.com/satimoto/go-datastore/pkg/location/mocks"
	sessionMocks "github.com/satimoto/go-datastore/pkg/session/mocks"
	countryaccount "github.com/satimoto/go-lsp/internal/countryaccount/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	"github.com/satimoto/go-lsp/internal/session"
	tariff "github.com/satimoto/go-lsp/internal/tariff/mocks"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *session.SessionResolver {
	return &session.SessionResolver{
		Repository:             sessionMocks.NewRepository(repositoryService),
		LightningService:       lightningService,
		NotificationService:    notificationService,
		OcpiService:            ocpiService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationRepository:     location.NewRepository(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolverWithServices(repositoryService, ocpiService),
	}
}
