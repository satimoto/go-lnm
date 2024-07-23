package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/session"
)

type RpcSessionResolver struct {
	SessionResolver *session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcSessionResolver {
	return &RpcSessionResolver{
		SessionResolver: session.NewResolver(repositoryService, services),
	}
}
