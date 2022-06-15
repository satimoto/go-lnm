package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/cdr"
	"github.com/satimoto/go-lsp/internal/ferp"
)

type RpcCdrResolver struct {
	CdrResolver *cdr.CdrResolver
	FerpService ferp.Ferp
}

func NewResolver(repositoryService *db.RepositoryService, ferpService ferp.Ferp) *RpcCdrResolver {
	return &RpcCdrResolver{
		CdrResolver: cdr.NewResolverWithFerp(repositoryService, ferpService),
		FerpService: ferpService,
	}
}
