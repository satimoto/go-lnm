package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	location "github.com/satimoto/go-lsp/internal/location/mocks"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/ocpi"
	"github.com/satimoto/go-lsp/internal/session"
	tariff "github.com/satimoto/go-lsp/internal/tariff/mocks"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *session.SessionResolver {
	repo := session.SessionRepository(repositoryService)

	return &session.SessionResolver{
		Repository:          repo,
		LightningService:    lightningnetwork.NewService(),
		NotificationService: notification.NewService(),
		OcpiService:         ocpi.NewService(),
		LocationResolver:    location.NewResolver(repositoryService),
		TariffResolver:      tariff.NewResolver(repositoryService),
		UserResolver:        user.NewResolver(repositoryService),
	}
}
