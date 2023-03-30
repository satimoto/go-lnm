package mocks

import (
	"context"

	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	node "github.com/satimoto/go-datastore/pkg/node/mocks"
	backup "github.com/satimoto/go-lnm/internal/backup/mocks"
	"github.com/satimoto/go-lnm/internal/monitor"
	channelbackup "github.com/satimoto/go-lnm/internal/monitor/channelbackup/mocks"
	htlcevent "github.com/satimoto/go-lnm/internal/monitor/htlcevent/mocks"
	invoice "github.com/satimoto/go-lnm/internal/monitor/invoice/mocks"
	transaction "github.com/satimoto/go-lnm/internal/monitor/transaction/mocks"
	"github.com/satimoto/go-lnm/internal/service"
)

func NewMonitor(shutdownCtx context.Context, repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *monitor.Monitor {
	backupService := backup.NewService()

	return &monitor.Monitor{
		LightningService:     services.LightningService,
		NodeRepository:       node.NewRepository(repositoryService),
		ChannelBackupMonitor: channelbackup.NewChannelBackupMonitor(repositoryService, backupService, services),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, services),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, services),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, services),
	}
}
