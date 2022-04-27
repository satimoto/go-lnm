package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/internal/user"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *user.UserResolver {
	repo := user.UserRepository(repositoryService)

	return &user.UserResolver{
		Repository: repo,
	}
}
