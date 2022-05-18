package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	countryaccount "github.com/satimoto/go-lsp/internal/countryaccount/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	location "github.com/satimoto/go-lsp/internal/location/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	"github.com/satimoto/go-lsp/internal/session"
	tariff "github.com/satimoto/go-lsp/internal/tariff/mocks"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *session.SessionResolver {
	repo := session.SessionRepository(repositoryService)

	return &session.SessionResolver{
		Repository:             repo,
		LightningService:       lightningService,
		NotificationService:    notificationService,
		OcpiService:            ocpiService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationResolver:       location.NewResolver(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolverWithServices(repositoryService, ocpiService),
	}
}
