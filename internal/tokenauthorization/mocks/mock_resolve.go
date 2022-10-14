package mocks

import (
	cdr "github.com/satimoto/go-datastore/pkg/cdr/mocks"
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	tokenauthorizationMocks "github.com/satimoto/go-datastore/pkg/tokenauthorization/mocks"
	"github.com/satimoto/go-lsp/internal/service"
	session "github.com/satimoto/go-lsp/internal/session/mocks"
	"github.com/satimoto/go-lsp/internal/tokenauthorization"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *tokenauthorization.TokenAuthorizationResolver {
	return &tokenauthorization.TokenAuthorizationResolver{
		Repository:          tokenauthorizationMocks.NewRepository(repositoryService),
		NotificationService: services.NotificationService,
		CdrRepository:       cdr.NewRepository(repositoryService),
		SessionResolver:     *session.NewResolver(repositoryService, services),
	}
}
