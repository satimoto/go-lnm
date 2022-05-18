package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/session"
)

type RpcSessionRepository interface{}

type RpcSessionResolver struct {
	Repository RpcSessionRepository
	*session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcSessionResolver {
	repo := RpcSessionRepository(repositoryService)
	return &RpcSessionResolver{
		Repository:      repo,
		SessionResolver: session.NewResolver(repositoryService),
	}
}
