package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/blockepoch"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewBlockEpochMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *blockepoch.BlockEpochMonitor {
	return &blockepoch.BlockEpochMonitor{
		LightningService: services.LightningService,
	}
}
