package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/internal/node"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *node.NodeResolver {
	repo := node.NodeRepository(repositoryService)

	return &node.NodeResolver{
		Repository: repo,
	}
}
