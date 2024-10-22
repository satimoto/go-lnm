package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	location "github.com/satimoto/go-datastore/pkg/location/mocks"
	sessionMocks "github.com/satimoto/go-datastore/pkg/session/mocks"
	tokenauthorization "github.com/satimoto/go-datastore/pkg/tokenauthorization/mocks"
	account "github.com/satimoto/go-lnm/internal/account/mocks"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/session"
	tariff "github.com/satimoto/go-lnm/internal/tariff/mocks"
	user "github.com/satimoto/go-lnm/internal/user/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *session.SessionResolver {
	return &session.SessionResolver{
		Repository:                   sessionMocks.NewRepository(repositoryService),
		FerpService:                  services.FerpService,
		LightningService:             services.LightningService,
		NotificationService:          services.NotificationService,
		OcpiService:                  services.OcpiService,
		AccountResolver:              account.NewResolver(repositoryService),
		LocationRepository:           location.NewRepository(repositoryService),
		TariffResolver:               tariff.NewResolver(repositoryService),
		TokenAuthorizationRepository: tokenauthorization.NewRepository(repositoryService),
		UserResolver:                 user.NewResolver(repositoryService, services),
	}
}
