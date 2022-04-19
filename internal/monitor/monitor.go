package monitor

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/monitor/channelevent"
	"github.com/satimoto/go-lsp/internal/monitor/custommessage"
	"github.com/satimoto/go-lsp/internal/monitor/htlc"
	"github.com/satimoto/go-lsp/internal/monitor/htlcevent"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
	"github.com/satimoto/go-lsp/internal/node"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Monitor struct {
	*grpc.ClientConn
	lnrpc.LightningClient
	MacaroonCtx context.Context

	*node.NodeResolver

	*channelevent.ChannelEventMonitor
	*custommessage.CustomMessageMonitor
	*htlc.HtlcMonitor
	*htlcevent.HtlcEventMonitor
	*transaction.TransactionMonitor
}

func NewMonitor(repositoryService *db.RepositoryService) *Monitor {
	lndTlsCert, err := base64.StdEncoding.DecodeString(os.Getenv("LND_TLS_CERT"))
	util.PanicOnError("Invalid LND TLS Certificate", err)

	credentials, err := util.NewCredential(string(lndTlsCert))
	util.PanicOnError("Error creating transport credentials", err)

	clientConn, err := grpc.Dial(os.Getenv("LND_GRPC_HOST"), grpc.WithTransportCredentials(credentials))
	util.PanicOnError("Error connecting to LND host", err)

	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	customMessageMonitor := custommessage.NewCustomMessageMonitor(repositoryService)

	return &Monitor{
		ClientConn:      clientConn,
		LightningClient: lnrpc.NewLightningClient(clientConn),
		MacaroonCtx:     macaroonCtx,

		NodeResolver: node.NewResolver(repositoryService),

		ChannelEventMonitor:  channelevent.NewChannelEventMonitor(repositoryService),
		CustomMessageMonitor: customMessageMonitor,
		HtlcMonitor:          htlc.NewHtlcMonitor(repositoryService, customMessageMonitor),
		HtlcEventMonitor:     htlcevent.NewHtlcEventMonitor(repositoryService),
		TransactionMonitor:   transaction.NewTransactionMonitor(repositoryService),
	}
}

func (m *Monitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	err := m.register()
	util.PanicOnError("Error registering LSP", err)

	m.ChannelEventMonitor.SetClientConnection(m.ClientConn, m.MacaroonCtx)
	m.CustomMessageMonitor.SetClientConnection(m.ClientConn, m.MacaroonCtx)
	m.HtlcMonitor.SetClientConnection(m.ClientConn, m.MacaroonCtx)
	m.HtlcEventMonitor.SetClientConnection(m.ClientConn, m.MacaroonCtx)
	m.TransactionMonitor.SetClientConnection(m.ClientConn, m.MacaroonCtx)

	m.ChannelEventMonitor.StartMonitor(ctx, waitGroup)
	m.CustomMessageMonitor.StartMonitor(ctx, waitGroup)
	m.HtlcMonitor.StartMonitor(ctx, waitGroup)
	m.HtlcEventMonitor.StartMonitor(ctx, waitGroup)
	m.TransactionMonitor.StartMonitor(ctx, waitGroup)
}

func (m *Monitor) register() error {
	waitingForSync := false

	for {
		getInfoResponse, err := m.LightningClient.GetInfo(m.MacaroonCtx, &lnrpc.GetInfoRequest{})

		if err != nil {
			log.Printf("Error getting info: %v", err)
			return err
		}

		ip, err := util.GetIPAddress()
		util.PanicOnError("Error getting IP address", err)

		if !waitingForSync {
			log.Print("Registering node")
			log.Printf("Version: %v", getInfoResponse.Version)
			log.Printf("CommitHash: %v", getInfoResponse.CommitHash)
			log.Printf("IdentityPubkey: %v", getInfoResponse.IdentityPubkey)
			log.Printf("IP Address: %v", ip.String())
		}

		if getInfoResponse.SyncedToChain {
			// Register node
			ctx := context.Background()
			numChannels := int64(getInfoResponse.NumActiveChannels + getInfoResponse.NumInactiveChannels + getInfoResponse.NumPendingChannels)
			numPeers := int64(getInfoResponse.NumPeers)
			addr := os.Getenv("LND_P2P_HOST")

			if n, err := m.NodeResolver.Repository.GetNodeByPubkey(ctx, getInfoResponse.IdentityPubkey); err == nil {
				// Update node
				updateNodeParams := node.NewUpdateNodeParams(n)
				updateNodeParams.NodeAddr = addr
				updateNodeParams.LspAddr = ip.String()
				updateNodeParams.Alias = getInfoResponse.Alias
				updateNodeParams.Color = getInfoResponse.Color
				updateNodeParams.CommitHash = getInfoResponse.CommitHash
				updateNodeParams.Version = getInfoResponse.Version
				updateNodeParams.Channels = numChannels
				updateNodeParams.Peers = numPeers

				m.NodeResolver.Repository.UpdateNode(ctx, updateNodeParams)
			} else {
				// Create node
				createNodeParams := db.CreateNodeParams{
					Pubkey:     getInfoResponse.IdentityPubkey,
					NodeAddr:   addr,
					LspAddr:    ip.String(),
					Alias:      getInfoResponse.Alias,
					Color:      getInfoResponse.Color,
					CommitHash: getInfoResponse.CommitHash,
					Version:    getInfoResponse.Version,
					Channels:   numChannels,
					Peers:      numPeers,
				}

				m.NodeResolver.Repository.CreateNode(ctx, createNodeParams)
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
