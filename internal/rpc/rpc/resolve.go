package rpc

import (
	"github.com/satimoto/go-datastore/pkg/db"
)

type RpcResolver struct{}

func NewResolver(repositoryService *db.RepositoryService) *RpcResolver {
	return &RpcResolver{}
}
