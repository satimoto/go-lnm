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
	peerevent "github.com/satimoto/go-lsp/internal/monitor/peerevent/mocks"
	psbtfund "github.com/satimoto/go-lsp/internal/monitor/psbtfund/mocks"
	transaction "github.com/satimoto/go-lsp/internal/monitor/transaction/mocks"
	notification "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpi "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, ferpService *ferp.MockFerpService, lightningService *lightningnetwork.MockLightningNetworkService, notificationService *notification.MockNotificationService, ocpiService *ocpi.MockOcpiService) *monitor.Monitor {
	backupService := backup.NewService()
	psbtFundService := psbtfund.NewService(repositoryService, lightningService)
	htlcMonitor := htlc.NewHtlcMonitor(repositoryService, lightningService, psbtFundService)

	return &monitor.Monitor{
		LightningService:     lightningService,
		PsbtFundService:      psbtFundService,
		NodeRepository:       node.NewRepository(repositoryService),
		ChannelBackupMonitor: channelbackup.NewChannelBackupMonitor(repositoryService, backupService, lightningService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService, htlcMonitor),
		CustomMessageMonitor: custommessage.NewCustomMessageMonitor(repositoryService, lightningService),
		HtlcMonitor:          htlcMonitor,
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, ferpService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, ferpService, lightningService, notificationService, ocpiService),
		PeerEventMonitor:     peerevent.NewPeerEventMonitor(repositoryService, lightningService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}
