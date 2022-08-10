package tariff

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/tariff"
)

type TariffResolver struct {
	Repository tariff.TariffRepository
}

func NewResolver(repositoryService *db.RepositoryService) *TariffResolver {
	return &TariffResolver{
		Repository: tariff.NewRepository(repositoryService),
	}
}
