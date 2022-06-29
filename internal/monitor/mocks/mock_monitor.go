package mocks

import (
	"context"

	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	backup "github.com/satimoto/go-lsp/internal/backup/mocks"
	ferp "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor"
	channelbackup "github.com/satimoto/go-lsp/internal/monitor/channelbackup/mocks"
	channelevent "github.com/satimoto/go-lsp/internal/monitor/channelevent/mocks"
	custommessage "github.com/satimoto/go-lsp/internal/monitor/custommessage/mocks"
	htlc "github.com/satimoto/go-lsp/internal/monitor/htlc/mocks"
	htlcevent "github.com/satimoto/go-lsp/internal/monitor/htlcevent/mocks"
	invoice "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	transaction "github.com/satimoto/go-lsp/internal/monitor/transaction/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *monitor.Monitor {
	backupService := backup.NewService()
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService, lightningService)

	return &monitor.Monitor{
		LightningService:     lightningService,
		ShutdownCtx:          shutdownCtx,
		NodeRepository:       node.NewRepository(repositoryService),
		ChannelBackupMonitor: channelbackup.NewChannelBackupMonitor(repositoryService, backupService, lightningService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, lightningService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, ferpService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, lightningService, notificationService, ocpiService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}
