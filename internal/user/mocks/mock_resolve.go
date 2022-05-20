package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	userMocks "github.com/satimoto/go-datastore/pkg/user/mocks"
	"github.com/satimoto/go-lsp/internal/user"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *user.UserResolver {
	return NewResolverWithServices(repositoryService, ocpi.NewService())
}

func NewResolverWithServices(repositoryService *mocks.MockRepositoryService, ocpiService *ocpi.MockOcpiService) *user.UserResolver {
	return &user.UserResolver{
		Repository:  userMocks.NewRepository(repositoryService),
		OcpiService: ocpiService,
	}
}
