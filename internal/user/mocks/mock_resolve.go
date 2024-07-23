package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	userMocks "github.com/satimoto/go-datastore/pkg/user/mocks"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/user"
)

func NewResolver(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *user.UserResolver {
	return &user.UserResolver{
		Repository:  userMocks.NewRepository(repositoryService),
		OcpiService: services.OcpiService,
	}
}
