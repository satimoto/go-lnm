package mocks

import (
	"context"

	mocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	ocpi "github.com/satimoto/go-lsp/internal/ocpi/mocks"
	"github.com/satimoto/go-lsp/internal/monitor"
	channelevent "github.com/satimoto/go-lsp/internal/monitor/channelevent/mocks"
	custommessage "github.com/satimoto/go-lsp/internal/monitor/custommessage/mocks"
	htlc "github.com/satimoto/go-lsp/internal/monitor/htlc/mocks"
	htlcevent "github.com/satimoto/go-lsp/internal/monitor/htlcevent/mocks"
	invoice "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	transaction "github.com/satimoto/go-lsp/internal/monitor/transaction/mocks"
	node "github.com/satimoto/go-lsp/internal/node/mocks"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, ocpiService *ocpi.MockOcpiService) *monitor.Monitor {
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService, lightningService)

	return &monitor.Monitor{
		LightningService:     lightningService,
		ShutdownCtx:          shutdownCtx,
		NodeResolver:         node.NewResolver(repositoryService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, lightningService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, lightningService, ocpiService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}
