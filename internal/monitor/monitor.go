package monitor

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/backup"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/monitor/blockepoch"
	"github.com/satimoto/go-lsp/internal/monitor/channelacceptor"
	"github.com/satimoto/go-lsp/internal/monitor/channelbackup"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/channelgraph"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	"github.com/satimoto/go-lsp/internal/monitor/peerevent"
	"github.com/satimoto/go-lsp/internal/monitor/pendingnotification"
	"github.com/satimoto/go-lsp/internal/monitor/psbtfund"
	"github.com/satimoto/go-lsp/internal/monitor/startup"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/pkg/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type Monitor struct {
	LightningService           lightningnetwork.LightningNetwork
	PsbtFundService            psbtfund.PsbtFund
	StartupService             startup.Startup
	NodeRepository             node.NodeRepository
	BlockEpochMonitor          *blockepoch.BlockEpochMonitor
	ChannelAcceptorMonitor     *channelacceptor.ChannelAcceptorMonitor
	ChannelBackupMonitor       *channelbackup.ChannelBackupMonitor
	ChannelEventMonitor        *channelevent.ChannelEventMonitor
	ChannelGraphMonitor        *channelgraph.ChannelGraphMonitor
	CustomMessageMonitor       *custommessage.CustomMessageMonitor
	HtlcMonitor                *htlc.HtlcMonitor
	HtlcEventMonitor           *htlcevent.HtlcEventMonitor
	InvoiceMonitor             *invoice.InvoiceMonitor
	PendingNotificationMonitor *pendingnotification.PendingNotificationMonitor
	PeerEventMonitor           *peerevent.PeerEventMonitor
	TransactionMonitor         *transaction.TransactionMonitor
	nodeID                     int64
	shutdownCtx                context.Context
}

func NewMonitor(shutdownCtx context.Context, repositoryService *db.RepositoryService, services *service.ServiceResolver) *Monitor {
	backupService := backup.NewService()
	psbtFundService := psbtfund.NewService(repositoryService, services)
	startupService := startup.NewService(repositoryService, services)
	htlcMonitor := htlc.NewHtlcMonitor(repositoryService, services, psbtFundService)

	return &Monitor{
		LightningService:           services.LightningService,
		PsbtFundService:            psbtFundService,
		StartupService:             startupService,
		NodeRepository:             node.NewRepository(repositoryService),
		BlockEpochMonitor:          blockepoch.NewBlockEpochMonitor(repositoryService, services),
		ChannelAcceptorMonitor:     channelacceptor.NewChannelAcceptorMonitor(repositoryService, services),
		ChannelBackupMonitor:       channelbackup.NewChannelBackupMonitor(repositoryService, backupService, services),
		ChannelEventMonitor:        channelevent.NewChannelEventMonitor(repositoryService, services, htlcMonitor),
		ChannelGraphMonitor:        channelgraph.NewChannelGraphMonitor(repositoryService, services, htlcMonitor),
		CustomMessageMonitor:       custommessage.NewCustomMessageMonitor(repositoryService, services),
		HtlcMonitor:                htlcMonitor,
		HtlcEventMonitor:           htlcevent.NewHtlcEventMonitor(repositoryService, services),
		InvoiceMonitor:             invoice.NewInvoiceMonitor(repositoryService, services),
		PeerEventMonitor:           peerevent.NewPeerEventMonitor(repositoryService, services),
		PendingNotificationMonitor: pendingnotification.NewPendingNotificationMonitor(repositoryService, services),
		TransactionMonitor:         transaction.NewTransactionMonitor(repositoryService, services),
		shutdownCtx:                shutdownCtx,
	}
}

func (m *Monitor) StartMonitor(waitGroup *sync.WaitGroup) {
	err := m.register()
	dbUtil.PanicOnError("LSP010", "Error registering LSP", err)

	m.PsbtFundService.Start(m.nodeID, m.shutdownCtx, waitGroup)
	m.StartupService.Start(m.nodeID, m.shutdownCtx, waitGroup)
	m.BlockEpochMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.ChannelAcceptorMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.ChannelBackupMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.ChannelEventMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.ChannelGraphMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.CustomMessageMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.HtlcMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.HtlcEventMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.InvoiceMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.PeerEventMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.PendingNotificationMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
	m.TransactionMonitor.StartMonitor(m.nodeID, m.shutdownCtx, waitGroup)
}

func (m *Monitor) register() error {
	ctx := context.Background()
	rpcHost := os.Getenv("RPC_HOST")
	waitingForSync := false

	if len(rpcHost) == 0 {
		ipAddr, err := util.GetIPAddress()
		dbUtil.PanicOnError("LSP011", "Error getting IP address", err)
		rpcHost = ipAddr
	}

	lspAddr := fmt.Sprintf("%s:%s", rpcHost, os.Getenv("RPC_PORT"))
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	_, err := ocpiService.TestConnection(ctx, &ocpirpc.TestConnectionRequest{
		Addr: lspAddr,
	})

	dbUtil.PanicOnError("LSP047", "Error testing RPC connectivity", err)

	for {
		getInfoResponse, err := m.LightningService.GetInfo(&lnrpc.GetInfoRequest{})

		if err != nil {
			metrics.RecordError("LSP004", "Error getting info", err)
			return err
		}

		if !waitingForSync {
			log.Print("Registering node")
			log.Printf("Version: %v", getInfoResponse.Version)
			log.Printf("CommitHash: %v", getInfoResponse.CommitHash)
			log.Printf("IdentityPubkey: %v", getInfoResponse.IdentityPubkey)
			log.Printf("LSP Address: %v", lspAddr)
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
				updateNodeParams.LspAddr = lspAddr
				updateNodeParams.Alias = getInfoResponse.Alias
				updateNodeParams.Color = getInfoResponse.Color
				updateNodeParams.CommitHash = getInfoResponse.CommitHash
				updateNodeParams.Version = getInfoResponse.Version
				updateNodeParams.Channels = numChannels
				updateNodeParams.Peers = numPeers

				updatedNode, err := m.NodeRepository.UpdateNode(ctx, updateNodeParams)

				if err != nil {
					metrics.RecordError("LSP075", "Error updating node", err)
					log.Printf("LSP075: Params=%#v", updateNodeParams)
				}

				m.nodeID = updatedNode.ID
			} else {
				// Create node
				createNodeParams := db.CreateNodeParams{
					Pubkey:     getInfoResponse.IdentityPubkey,
					NodeAddr:   lndAddr,
					LspAddr:    lspAddr,
					Alias:      getInfoResponse.Alias,
					Color:      getInfoResponse.Color,
					CommitHash: getInfoResponse.CommitHash,
					Version:    getInfoResponse.Version,
					Channels:   numChannels,
					Peers:      numPeers,
					IsActive:   true,
				}

				createdNode, err := m.NodeRepository.CreateNode(ctx, createNodeParams)

				if err != nil {
					metrics.RecordError("LSP076", "Error creating node", err)
					log.Printf("LSP076: Params=%#v", createNodeParams)
				}

				m.nodeID = createdNode.ID
			}

			channelevent.RecordChannels(uint32(numChannels))

			log.Print("Registered node")

			estimateFeeResponse, err := m.LightningService.EstimateFee(&walletrpc.EstimateFeeRequest{ConfTarget: 6})
			
			if err != nil {
				metrics.RecordError("LSP154", "Error getting fee estimation", err)
				break
			}

			log.Printf("Startup fee estimation: %v", estimateFeeResponse.SatPerKw)
			break
		}

		waitingForSync = true
		log.Printf("BlockHeight: %v", getInfoResponse.BlockHeight)
		log.Printf("BestHeaderTimestamp: %v", getInfoResponse.BestHeaderTimestamp)
		time.Sleep(6 * time.Second)
	}

	return nil
}

func getClearnetUri(uris []string) *string {
	for _, uri := range uris {
		if !strings.Contains(uri, "onion") {
			return &uri
		}
	}

	return nil
}
