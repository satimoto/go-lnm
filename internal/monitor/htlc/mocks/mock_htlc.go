package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	psbtfund "github.com/satimoto/go-lsp/internal/monitor/psbtfund/mocks"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewHtlcMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver, psbtFundService *psbtfund.MockPsbtFundService) *htlc.HtlcMonitor {
	return &htlc.HtlcMonitor{
		LightningService:       services.LightningService,
		PsbtFundService:        psbtFundService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
