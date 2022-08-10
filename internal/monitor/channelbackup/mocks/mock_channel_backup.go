package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	backup "github.com/satimoto/go-lsp/internal/backup/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/channelbackup"
)

func NewChannelBackupMonitor(repositoryService *mocks.MockRepositoryService, backupService *backup.MockBackupService, lightningService *lightningnetwork.MockLightningNetworkService) *channelbackup.ChannelBackupMonitor {
	return &channelbackup.ChannelBackupMonitor{
		BackupService:    backupService,
		LightningService: lightningService,
	}
}
