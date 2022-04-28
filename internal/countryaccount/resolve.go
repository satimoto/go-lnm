package countryaccount

import (
	"context"

	"github.com/satimoto/go-datastore/db"
)

type CountryAccountRepository interface {
	GetCountryAccountByCountry(ctx context.Context, country string) (db.CountryAccount, error)
}

type CountryAccountResolver struct {
	Repository CountryAccountRepository
}

func NewResolver(repositoryService *db.RepositoryService) *CountryAccountResolver {
	repo := CountryAccountRepository(repositoryService)
	return &CountryAccountResolver{repo}
}
