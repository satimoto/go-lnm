package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	backup "github.com/satimoto/go-lnm/internal/backup/mocks"
	"github.com/satimoto/go-lnm/internal/monitor/channelbackup"
	"github.com/satimoto/go-lnm/internal/service"
)

func NewChannelBackupMonitor(repositoryService *mocks.MockRepositoryService, backupService *backup.MockBackupService, services *service.ServiceResolver) *channelbackup.ChannelBackupMonitor {
	return &channelbackup.ChannelBackupMonitor{
		BackupService:    backupService,
		LightningService: services.LightningService,
	}
}
