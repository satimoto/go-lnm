package monitor

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/backup"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/monitor/blockepoch"
	"github.com/satimoto/go-lnm/internal/monitor/channelbackup"
	"github.com/satimoto/go-lnm/internal/monitor/htlcevent"
	"github.com/satimoto/go-lnm/internal/monitor/invoice"
	"github.com/satimoto/go-lnm/internal/monitor/pendingnotification"
	"github.com/satimoto/go-lnm/internal/monitor/startup"
	"github.com/satimoto/go-lnm/internal/monitor/transaction"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/pkg/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type Monitor struct {
	LightningService           lightningnetwork.LightningNetwork
	StartupService             startup.Startup
	NodeRepository             node.NodeRepository
	BlockEpochMonitor          *blockepoch.BlockEpochMonitor
	ChannelBackupMonitor       *channelbackup.ChannelBackupMonitor
	HtlcEventMonitor           *htlcevent.HtlcEventMonitor
	InvoiceMonitor             *invoice.InvoiceMonitor
	PendingNotificationMonitor *pendingnotification.PendingNotificationMonitor
	TransactionMonitor         *transaction.TransactionMonitor
	nodeID                     int64
	shutdownCtx                context.Context
}

func NewMonitor(shutdownCtx context.Context, repositoryService *db.RepositoryService, services *service.ServiceResolver) *Monitor {
	backupService := backup.NewService()
	startupService := startup.NewService(repositoryService, services)

	return &Monitor{
		LightningService:           services.LightningService,
		StartupService:             startupService,
		NodeRepository:             node.NewRepository(repositoryService),
		BlockEpochMonitor:          blockepoch.NewBlockEpochMonitor(repositoryService, services),
		ChannelBackupMonitor:       channelbackup.NewChannelBackupMonitor(repositoryService, backupService, services),
		HtlcEventMonitor:           htlcevent.NewHtlcEventMonitor(repositoryService, services),
		InvoiceMonitor:             invoice.NewInvoiceMonitor(repositoryService, services),
		PendingNotificationMonitor: pendingnotification.NewPendingNotificationMonitor(repositoryService, services),
		TransactionMonitor:         transaction.NewTransactionMonitor(repositoryService, services),
		shutdownCtx:                shutdownCtx,
	}
}

func (m *Monitor) StartMonitor(waitGroup *sync.WaitGroup) {
	err := m.register()
	dbUtil.PanicOnError("LNM010", "Error registering LSP", err)

	m.StartupService.Start(m.nodeID, m.shutdownCtx, waitGroup)
	m.BlockEpochMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.ChannelBackupMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.HtlcEventMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.InvoiceMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.PendingNotificationMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.TransactionMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
}

func (m *Monitor) register() error {
	ctx := context.Background()
	rpcHost := os.Getenv("RPC_HOST")
	testRpcConnection := dbUtil.GetEnvBool("TEST_RPC_CONNECTION", true)
	waitingForSync := false

	if len(rpcHost) == 0 {
		ipAddr, err := util.GetIPAddress()
		dbUtil.PanicOnError("LNM011", "Error getting IP address", err)
		rpcHost = ipAddr
	}

	rpcAddr := fmt.Sprintf("%s:%s", rpcHost, os.Getenv("RPC_PORT"))

	if testRpcConnection {
		ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

		_, err := ocpiService.TestConnection(ctx, &ocpirpc.TestConnectionRequest{
			Addr: rpcAddr,
		})

		dbUtil.PanicOnError("LNM047", "Error testing RPC connectivity", err)
	}

	for {
		getInfoResponse, err := m.LightningService.GetInfo(&lnrpc.GetInfoRequest{})

		if err != nil {
			metrics.RecordError("LNM004", "Error getting info", err)
			return err
		}

		if !waitingForSync {
			log.Print("Registering node")
			log.Printf("Version: %v", getInfoResponse.Version)
			log.Printf("CommitHash: %v", getInfoResponse.CommitHash)
			log.Printf("IdentityPubkey: %v", getInfoResponse.IdentityPubkey)
			log.Printf("RPC Address: %v", rpcAddr)
		}

		if getInfoResponse.SyncedToChain {
			// Register node
			numChannels := int64(getInfoResponse.NumActiveChannels + getInfoResponse.NumInactiveChannels + getInfoResponse.NumPendingChannels)
			numPeers := int64(getInfoResponse.NumPeers)
			lightningAddr := util.NewLightningAddr(getInfoResponse.Uris[0])
			lndAddr := dbUtil.GetEnv("LND_P2P_HOST", lightningAddr.Host)

			if n, err := m.NodeRepository.GetNodeByPubkey(ctx, getInfoResponse.IdentityPubkey); err == nil {
				// Update node
				updateNodeParams := param.NewUpdateNodeParams(n)
				updateNodeParams.NodeAddr = lndAddr
				updateNodeParams.RpcAddr = rpcAddr
				updateNodeParams.Alias = getInfoResponse.Alias
				updateNodeParams.Color = getInfoResponse.Color
				updateNodeParams.CommitHash = getInfoResponse.CommitHash
				updateNodeParams.Version = getInfoResponse.Version
				updateNodeParams.Channels = numChannels
				updateNodeParams.Peers = numPeers

				updatedNode, err := m.NodeRepository.UpdateNode(ctx, updateNodeParams)

				if err != nil {
					metrics.RecordError("LNM075", "Error updating node", err)
					log.Printf("LNM075: Params=%#v", updateNodeParams)
				}

				m.nodeID = updatedNode.ID
			} else {
				// Create node
				createNodeParams := db.CreateNodeParams{
					Pubkey:     getInfoResponse.IdentityPubkey,
					NodeAddr:   lndAddr,
					RpcAddr:    rpcAddr,
					Alias:      getInfoResponse.Alias,
					Color:      getInfoResponse.Color,
					CommitHash: getInfoResponse.CommitHash,
					Version:    getInfoResponse.Version,
					Channels:   numChannels,
					Peers:      numPeers,
					IsActive:   true,
					IsLsp:      false,
				}

				createdNode, err := m.NodeRepository.CreateNode(ctx, createNodeParams)

				if err != nil {
					metrics.RecordError("LNM076", "Error creating node", err)
					log.Printf("LNM076: Params=%#v", createNodeParams)
				}

				m.nodeID = createdNode.ID
			}

			log.Print("Registered node")
			break
		}

		waitingForSync = true
		log.Printf("BlockHeight: %v", getInfoResponse.BlockHeight)
		log.Printf("BestHeaderTimestamp: %v", getInfoResponse.BestHeaderTimestamp)
		time.Sleep(6 * time.Second)
	}

	return nil
}
