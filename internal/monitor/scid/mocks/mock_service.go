package mocks

import (
	"sync"

	"github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/scid"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewService(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) scid.Scid {
	return &scid.ScidService{
		LightningService: services.LightningService,
		NodeRepository:   node.NewRepository(repositoryService),
		Mutex:            &sync.Mutex{},
	}
}