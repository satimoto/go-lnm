package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/cdr"
)

type RpcCdrRepository interface{}

type RpcCdrResolver struct {
	Repository RpcCdrRepository
	*cdr.CdrResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcCdrResolver {
	repo := RpcCdrRepository(repositoryService)
	return &RpcCdrResolver{
		Repository:  repo,
		CdrResolver: cdr.NewResolver(repositoryService),
	}
}
