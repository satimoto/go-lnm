package user

import (
	"context"

	"github.com/satimoto/go-datastore/db"
)

type UserRepository interface {
	GetUserByTokenID(ctx context.Context, id int64) (db.User, error)
}

type UserResolver struct {
	Repository UserRepository
}

func NewResolver(repositoryService *db.RepositoryService) *UserResolver {
	repo := UserRepository(repositoryService)
	return &UserResolver{repo}
}
