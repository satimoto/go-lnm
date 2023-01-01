package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	userMocks "github.com/satimoto/go-datastore/pkg/user/mocks"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *user.UserResolver {
	return &user.UserResolver{
		Repository:  userMocks.NewRepository(repositoryService),
		OcpiService: services.OcpiService,
	}
}
