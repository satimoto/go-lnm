package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewCustomMessageMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *custommessage.CustomMessageMonitor {
	return &custommessage.CustomMessageMonitor{
		LightningService:      services.LightningService,
		CustomMessageHandlers: make(map[string]custommessage.CustomMessageHandler),
	}
}
