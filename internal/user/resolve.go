package user

import (
	"context"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type UserRepository interface {
	GetUser(ctx context.Context, id int64) (db.User, error)
	GetUserBySessionID(ctx context.Context, id int64) (db.User, error)
	GetUserByTokenID(ctx context.Context, id int64) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
}

type UserResolver struct {
	Repository  UserRepository
	OcpiService ocpi.Ocpi
}

func NewResolver(repositoryService *db.RepositoryService) *UserResolver {
	return NewResolverWithServices(repositoryService, ocpi.NewService())
}

func NewResolverWithServices(repositoryService *db.RepositoryService, ocpiService ocpi.Ocpi) *UserResolver {
	repo := UserRepository(repositoryService)

	return &UserResolver{
		Repository:  repo,
		OcpiService: ocpiService,
	}
}
