package channelbackup

import (
	"context"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/backup"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelBackupMonitor struct {
	LightningService     lightningnetwork.LightningNetwork
	BackupService        backup.Backup
	ChannelBackupsClient lnrpc.Lightning_SubscribeChannelBackupsClient
	nodeID               int64
}

func NewChannelBackupMonitor(repositoryService *db.RepositoryService, backupService backup.Backup, services *service.ServiceResolver) *ChannelBackupMonitor {
	return &ChannelBackupMonitor{
		BackupService:    backupService,
		LightningService: services.LightningService,
	}
}

func (m *ChannelBackupMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Channel Backups")
	channelBackupChan := make(chan lnrpc.ChanBackupSnapshot)

	m.nodeID = nodeID
	go m.waitForChannelBackups(shutdownCtx, waitGroup, channelBackupChan)
	go m.subscribeChannelBackupInterceptions(channelBackupChan)
}

func (m *ChannelBackupMonitor) handleChannelBackup(channelBackup lnrpc.ChanBackupSnapshot) {
	/** Channel Backup received.
	 *
	 */

	log.Print("Channel Backup")
	log.Printf("MultiChanBackup: %v", hex.EncodeToString(channelBackup.MultiChanBackup.MultiChanBackup))

	go m.BackupService.BackupChannelsWithRetry(channelBackup.MultiChanBackup.MultiChanBackup, 10)
}

func (m *ChannelBackupMonitor) subscribeChannelBackupInterceptions(channelBackupChan chan<- lnrpc.ChanBackupSnapshot) {
	htlcEventsClient, err := m.waitForSubscribeChannelBackupsClient(0, 1000)
	util.PanicOnError("LSP062", "Error creating Channel Backups client", err)
	m.ChannelBackupsClient = htlcEventsClient

	for {
		htlcInterceptRequest, err := m.ChannelBackupsClient.Recv()

		if err == nil {
			channelBackupChan <- *htlcInterceptRequest
		} else {
			m.ChannelBackupsClient, err = m.waitForSubscribeChannelBackupsClient(100, 1000)
			util.PanicOnError("LSP063", "Error creating Channel Backups client", err)
		}
	}
}

func (m *ChannelBackupMonitor) waitForChannelBackups(shutdownCtx context.Context, waitGroup *sync.WaitGroup, channelBackupChan chan lnrpc.ChanBackupSnapshot) {
	waitGroup.Add(1)
	defer close(channelBackupChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Channel Backups")
			return
		case htlcInterceptRequest := <-channelBackupChan:
			m.handleChannelBackup(htlcInterceptRequest)
		}
	}
}

func (m *ChannelBackupMonitor) waitForSubscribeChannelBackupsClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeChannelBackupsClient, err := m.LightningService.SubscribeChannelBackups(&lnrpc.ChannelBackupSubscription{})

		if err == nil {
			return subscribeChannelBackupsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Channel Backups client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
