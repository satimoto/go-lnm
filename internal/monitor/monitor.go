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
	"github.com/satimoto/go-lsp/internal/exchange"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
	"github.com/satimoto/go-lsp/internal/monitor/invoice"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
	"github.com/satimoto/go-lsp/internal/util"
	"github.com/satimoto/go-ocpi-api/ocpirpc"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type Monitor struct {
	LightningService     lightningnetwork.LightningNetwork
	ShutdownCtx          context.Context
	NodeRepository       node.NodeRepository
	ChannelEventMonitor  *channelevent.ChannelEventMonitor
	CustomMessageMonitor *custommessage.CustomMessageMonitor
	HtlcMonitor          *htlc.HtlcMonitor
	HtlcEventMonitor     *htlcevent.HtlcEventMonitor
	InvoiceMonitor       *invoice.InvoiceMonitor
	TransactionMonitor   *transaction.TransactionMonitor
}

func NewMonitor(shutdownCtx context.Context, repositoryService *db.RepositoryService, exchangeService exchange.Exchange) *Monitor {
	lightningService := lightningnetwork.NewService()
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService, lightningService)

	return &Monitor{
		LightningService:     lightningService,
		ShutdownCtx:          shutdownCtx,
		NodeRepository:       node.NewRepository(repositoryService),
		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService, lightningService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, lightningService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService, lightningService),
		InvoiceMonitor:       invoice.NewInvoiceMonitor(repositoryService, exchangeService, lightningService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService, lightningService),
	}
}

func (m *Monitor) StartMonitor(waitGroup *sync.WaitGroup) {
	err := m.register()
	dbUtil.PanicOnError("LSP010", "Error registering LSP", err)

	m.ChannelEventMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
	m.CustomMessageMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
	m.HtlcMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
	m.HtlcEventMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
	m.InvoiceMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
	m.TransactionMonitor.StartMonitor(m.ShutdownCtx, waitGroup)
}

func (m *Monitor) register() error {
	waitingForSync := false
	rpcHost := os.Getenv("RPC_HOST")

	if len(rpcHost) == 0 {
		ipAddr, err := util.GetIPAddress()
		dbUtil.PanicOnError("LSP011", "Error getting IP address", err)
		rpcHost = ipAddr
	}

	lspAddr := fmt.Sprintf("%s:%s", rpcHost, os.Getenv("RPC_PORT"))

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
			// Test RPC connectivity
			ctx := context.Background()
			ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

			_, err := ocpiService.TestConnection(ctx, &ocpirpc.TestConnectionRequest{
				Addr: lspAddr,
			})

			dbUtil.PanicOnError("LSP047", "Error testing RPC connectivity", err)
			
			// Register node
			numChannels := int64(getInfoResponse.NumActiveChannels + getInfoResponse.NumInactiveChannels + getInfoResponse.NumPendingChannels)
			numPeers := int64(getInfoResponse.NumPeers)
			lightningAddr := util.NewLightingAddr(getInfoResponse.Uris[0])

			if n, err := m.NodeRepository.GetNodeByPubkey(ctx, getInfoResponse.IdentityPubkey); err == nil {
				// Update node
				updateNodeParams := param.NewUpdateNodeParams(n)
				updateNodeParams.NodeAddr = lightningAddr.Host
				updateNodeParams.LspAddr = lspAddr
				updateNodeParams.Alias = getInfoResponse.Alias
				updateNodeParams.Color = getInfoResponse.Color
				updateNodeParams.CommitHash = getInfoResponse.CommitHash
				updateNodeParams.Version = getInfoResponse.Version
				updateNodeParams.Channels = numChannels
				updateNodeParams.Peers = numPeers

				m.NodeRepository.UpdateNode(ctx, updateNodeParams)
			} else {
				// Create node
				createNodeParams := db.CreateNodeParams{
					Pubkey:     getInfoResponse.IdentityPubkey,
					NodeAddr:   lightningAddr.Host,
					LspAddr:    lspAddr,
					Alias:      getInfoResponse.Alias,
					Color:      getInfoResponse.Color,
					CommitHash: getInfoResponse.CommitHash,
					Version:    getInfoResponse.Version,
					Channels:   numChannels,
					Peers:      numPeers,
				}

				m.NodeRepository.CreateNode(ctx, createNodeParams)
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
