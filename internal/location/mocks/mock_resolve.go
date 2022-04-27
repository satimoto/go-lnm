package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/internal/location"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *location.LocationResolver {
	repo := location.LocationRepository(repositoryService)

	return &location.LocationResolver{
		Repository: repo,
	}
}
