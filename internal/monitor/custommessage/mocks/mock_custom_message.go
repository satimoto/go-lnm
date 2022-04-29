package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
)

func NewCustomMessageMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *custommessage.CustomMessageMonitor {
	return &custommessage.CustomMessageMonitor{
		LightningService:      lightningService,
		CustomMessageHandlers: make(map[string]custommessage.CustomMessageHandler),
	}
}
