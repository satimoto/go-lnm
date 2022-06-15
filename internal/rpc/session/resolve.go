package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/ferp"
	"github.com/satimoto/go-lsp/internal/session"
)

type RpcSessionResolver struct {
	SessionResolver *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, ferpService ferp.Ferp) *RpcSessionResolver {
	return &RpcSessionResolver{
		SessionResolver: session.NewResolverWithFerp(repositoryService, ferpService),
	}
}
