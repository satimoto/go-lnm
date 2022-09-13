package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	psbtfund "github.com/satimoto/go-lsp/internal/monitor/psbtfund/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
)

func NewHtlcMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, psbtFundService *psbtfund.MockPsbtFundService) *htlc.HtlcMonitor {
	return &htlc.HtlcMonitor{
		LightningService:       lightningService,
		PsbtFundService:        psbtFundService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
