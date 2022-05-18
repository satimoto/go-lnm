package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/countryaccount"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *countryaccount.CountryAccountResolver {
	repo := countryaccount.CountryAccountRepository(repositoryService)

	return &countryaccount.CountryAccountResolver{
		Repository: repo,
	}
}
