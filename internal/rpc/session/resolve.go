package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/exchange"
	"github.com/satimoto/go-lsp/internal/session"
)

type RpcSessionResolver struct {
	SessionResolver *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, exchangeService exchange.Exchange) *RpcSessionResolver {
	return &RpcSessionResolver{
		SessionResolver: session.NewResolver(repositoryService, exchangeService),
	}
}
