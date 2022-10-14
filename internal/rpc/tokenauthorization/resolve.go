package tokenauthorization

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/internal/tokenauthorization"
)

type RpcTokenAuthorizationResolver struct {
	TokenAuthorizationResolver *tokenauthorization.TokenAuthorizationResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcTokenAuthorizationResolver {
	return &RpcTokenAuthorizationResolver{
		TokenAuthorizationResolver: tokenauthorization.NewResolver(repositoryService, services),
	}
}
