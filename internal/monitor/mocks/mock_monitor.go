package mocks

import (
	"context"

	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	backup "github.com/satimoto/go-lsp/internal/backup/mocks"
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
	"github.com/satimoto/go-lsp/internal/service"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *monitor.Monitor {
	backupService := backup.NewService()
	psbtFundService := psbtfund.NewService(repositoryService, services)
	htlcMonitor := htlc.NewHtlcMonitor(repositoryService, services, psbtFundService)

	return &monitor.Monitor{
		LightningService:     services.LightningService,
		PsbtFundService:      psbtFundService,
		NodeRepository:       node.NewRepository(repositoryService),
		ChannelBackupMonitor: channelbackup.NewChannelBackupMonitor(repositoryService, backupService, services),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, services, htlcMonitor),
		CustomMessageMonitor: custommessage.NewCustomMessageMonitor(repositoryService, services),
		HtlcMonitor:          htlcMonitor,
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, services),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, services),
		PeerEventMonitor:     peerevent.NewPeerEventMonitor(repositoryService, services),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, services),
	}
}
