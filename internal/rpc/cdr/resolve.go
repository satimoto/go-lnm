package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/cdr"
)

type RpcCdrResolver struct {
	CdrResolver *cdr.CdrResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcCdrResolver {
	return &RpcCdrResolver{
		CdrResolver: cdr.NewResolver(repositoryService),
	}
}
