package mocks

import (
	accountMocks "github.com/satimoto/go-datastore/pkg/account/mocks"
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/account"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *account.AccountResolver {
	return &account.AccountResolver{
		Repository: accountMocks.NewRepository(repositoryService),
	}
}
