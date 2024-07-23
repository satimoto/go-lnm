package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lnm/internal/cdr"
	"github.com/satimoto/go-lnm/internal/ferp"
	"github.com/satimoto/go-lnm/internal/service"
)

type RpcCdrResolver struct {
	CdrResolver *cdr.CdrResolver
	FerpService ferp.Ferp
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcCdrResolver {
	return &RpcCdrResolver{
		CdrResolver: cdr.NewResolver(repositoryService, services),
		FerpService: services.FerpService,
	}
}
