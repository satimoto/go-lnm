package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/blockepoch"
)

func NewBlockEpochMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *blockepoch.BlockEpochMonitor {
	return &blockepoch.BlockEpochMonitor{
		LightningService: lightningService,
	}
}
