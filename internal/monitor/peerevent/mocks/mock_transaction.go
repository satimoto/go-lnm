package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/peerevent"
	user "github.com/satimoto/go-lsp/internal/user/mocks"
)

func NewPeerEventMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *peerevent.PeerEventMonitor {
	return &peerevent.PeerEventMonitor{
		LightningService: lightningService,
		UserResolver:     user.NewResolver(repositoryService),
	}
}
