package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/tariff"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *tariff.TariffResolver {
	repo := tariff.TariffRepository(repositoryService)

	return &tariff.TariffResolver{
		Repository: repo,
	}
}
