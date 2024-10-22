package rpc

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lnm/internal/service"
)

type RpcResolver struct{}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcResolver {
	return &RpcResolver{}
}
