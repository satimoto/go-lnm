package user

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/user"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type UserResolver struct {
	Repository  user.UserRepository
	OcpiService ocpi.Ocpi
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *UserResolver {
	return &UserResolver{
		Repository:  user.NewRepository(repositoryService),
		OcpiService: services.OcpiService,
	}
}
