package user

import (
	"os"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/user"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type UserResolver struct {
	Repository  user.UserRepository
	OcpiService ocpi.Ocpi
}

func NewResolver(repositoryService *db.RepositoryService) *UserResolver {
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, ocpiService ocpi.Ocpi) *UserResolver {
	return &UserResolver{
		Repository:  user.NewRepository(repositoryService),
		OcpiService: ocpiService,
	}
}
