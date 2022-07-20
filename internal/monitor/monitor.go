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
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/backup"
	"github.com/satimoto/go-lsp/internal/ferp"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/monitor/blockepoch"
	"github.com/satimoto/go-lsp/internal/monitor/channelbackup"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
	"github.com/satimoto/go-lsp/internal/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type Monitor struct {
	LightningService     lightningnetwork.LightningNetwork
	ShutdownCtx          context.Context
	NodeRepository       node.NodeRepository
	BlockEpochMonitor    *blockepoch.BlockEpochMonitor
	ChannelBackupMonitor *channelbackup.ChannelBackupMonitor
	ChannelEventMonitor  *channelevent.ChannelEventMonitor
	CustomMessageMonitor *custommessage.CustomMessageMonitor
	HtlcMonitor          *htlc.HtlcMonitor
	HtlcEventMonitor     *htlcevent.HtlcEventMonitor
	InvoiceMonitor       *invoice.InvoiceMonitor
	TransactionMonitor   *transaction.TransactionMonitor
	nodeID               int64
}

func NewMonitor(shutdownCtx context.Context, repositoryService *db.RepositoryService, ferpService ferp.Ferp) *Monitor {
	backupService := backup.NewService()
	lightningService := lightningnetwork.NewService()
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService, lightningService)

	return &Monitor{
		LightningService:     lightningService,
		ShutdownCtx:          shutdownCtx,
		NodeRepository:       node.NewRepository(repositoryService),
		BlockEpochMonitor:    blockepoch.NewBlockEpochMonitor(repositoryService, lightningService),
		ChannelBackupMonitor: channelbackup.NewChannelBackupMonitor(repositoryService, backupService, lightningService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, lightningService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, ferpService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, ferpService, lightningService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}

func (m *Monitor) StartMonitor(waitGroup *sync.WaitGroup) {
	err := m.register()
	dbUtil.PanicOnError("LSP010", "Error registering LSP", err)

	m.BlockEpochMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.ChannelBackupMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.ChannelEventMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.CustomMessageMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.HtlcMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.HtlcEventMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.InvoiceMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
	m.TransactionMonitor.StartMonitor(m.nodeID, m.ShutdownCtx, waitGroup)
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
			dbUtil.LogOnError("LSP004", "Error getting info", err)
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
					dbUtil.LogOnError("LSP075", "Error updating node", err)
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
				}

				createdNode, err := m.NodeRepository.CreateNode(ctx, createNodeParams)

				if err != nil {
					dbUtil.LogOnError("LSP076", "Error creating node", err)
					log.Printf("LSP076: Params=%#v", createNodeParams)
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

func getClearnetUri(uris []string) *string {
	for _, uri := range uris {
		if !strings.Contains(uri, "onion") {
			return &uri
		}
	}

	return nil
}
