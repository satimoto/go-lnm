package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	countryaccount "github.com/satimoto/go-lsp/internal/countryaccount/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	location "github.com/satimoto/go-lsp/internal/location/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpi "github.com/satimoto/go-lsp/internal/ocpi/mocks"
	"github.com/satimoto/go-lsp/internal/session"
	tariff "github.com/satimoto/go-lsp/internal/tariff/mocks"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, ocpiService *ocpi.MockOcpiService) *session.SessionResolver {
	repo := session.SessionRepository(repositoryService)

	return &session.SessionResolver{
		Repository:             repo,
		LightningService:       lightningService,
		NotificationService:    notification.NewService(),
		OcpiService:            ocpiService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationResolver:       location.NewResolver(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolver(repositoryService),
	}
}
