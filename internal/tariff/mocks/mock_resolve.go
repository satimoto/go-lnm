package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	tariffMocks "github.com/satimoto/go-datastore/pkg/tariff/mocks"
	"github.com/satimoto/go-lnm/internal/tariff"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *tariff.TariffResolver {
	return &tariff.TariffResolver{
		Repository: tariffMocks.NewRepository(repositoryService),
	}
}
