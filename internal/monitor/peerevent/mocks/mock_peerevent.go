package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/peerevent"
	"github.com/satimoto/go-lsp/internal/service"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewPeerEventMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *peerevent.PeerEventMonitor {
	return &peerevent.PeerEventMonitor{
		LightningService: services.LightningService,
		UserResolver:     user.NewResolver(repositoryService, services),
	}
}
