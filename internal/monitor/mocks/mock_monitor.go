package mocks

import (
	"context"

	mocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor"
	channelevent "github.com/satimoto/go-lsp/internal/monitor/channelevent/mocks"
	custommessage "github.com/satimoto/go-lsp/internal/monitor/custommessage/mocks"
	htlc "github.com/satimoto/go-lsp/internal/monitor/htlc/mocks"
	htlcevent "github.com/satimoto/go-lsp/internal/monitor/htlcevent/mocks"
	invoice "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	transaction "github.com/satimoto/go-lsp/internal/monitor/transaction/mocks"
	node "github.com/satimoto/go-lsp/internal/node/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpi "github.com/satimoto/go-ocpi-api/pkg/ocpi/mocks"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *monitor.Monitor {
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService, lightningService)

	return &monitor.Monitor{
		LightningService:     lightningService,
		ShutdownCtx:          shutdownCtx,
		NodeResolver:         node.NewResolver(repositoryService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, lightningService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, lightningService, notificationService, ocpiService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}
