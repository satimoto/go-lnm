package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
	"github.com/satimoto/go-lsp/internal/user"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *user.UserResolver {
	return NewResolverWithServices(repositoryService, ocpi.NewService())
}

func NewResolverWithServices(repositoryService *mocks.MockRepositoryService, ocpiService *ocpi.MockOcpiService) *user.UserResolver {
	repo := user.UserRepository(repositoryService)

	return &user.UserResolver{
		Repository:  repo,
		OcpiService: ocpiService,
	}
}
