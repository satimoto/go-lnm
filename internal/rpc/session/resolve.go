package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/session"
)

type RpcSessionResolver struct {
	SessionResolver *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcSessionResolver {
	return &RpcSessionResolver{
		SessionResolver: session.NewResolver(repositoryService),
	}
}
