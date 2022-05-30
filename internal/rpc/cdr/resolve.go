package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/cdr"
	"github.com/satimoto/go-lsp/internal/exchange"
)

type RpcCdrResolver struct {
	CdrResolver     *cdr.CdrResolver
	ExchangeService exchange.Exchange
}

func NewResolver(repositoryService *db.RepositoryService, exchangeService exchange.Exchange) *RpcCdrResolver {
	return &RpcCdrResolver{
		CdrResolver:     cdr.NewResolver(repositoryService, exchangeService),
		ExchangeService: exchangeService,
	}
}
