package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/node"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/metadata"
)

func (i *Intercept) Register() error {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	waitingForSync := false

	for {
		getInfoResponse, err := i.LightningClient.GetInfo(macaroonCtx, &lnrpc.GetInfoRequest{})
		if err != nil {
			log.Printf("Error getting info: %v", err)
			return err
		}

		if !waitingForSync {
			log.Print("Registering node")
			log.Printf("Version: %v", getInfoResponse.Version)
			log.Printf("CommitHash: %v", getInfoResponse.CommitHash)
			log.Printf("IdentityPubkey: %v", getInfoResponse.IdentityPubkey)
		}

		if getInfoResponse.SyncedToChain {
			// Register node
			ctx := context.Background()
			numChannels := int64(getInfoResponse.NumActiveChannels + getInfoResponse.NumInactiveChannels + getInfoResponse.NumPendingChannels)
			numPeers := int64(getInfoResponse.NumPeers)
			addr := os.Getenv("LND_P2P_HOST")

			if n, err := i.NodeResolver.Repository.GetNodeByPubkey(ctx, getInfoResponse.IdentityPubkey); err == nil {
				// Update node
				updateNodeParams := node.NewUpdateNodeParams(n)
				updateNodeParams.Addr = addr
				updateNodeParams.Alias = getInfoResponse.Alias
				updateNodeParams.Color = getInfoResponse.Color
				updateNodeParams.CommitHash = getInfoResponse.CommitHash
				updateNodeParams.Version = getInfoResponse.Version
				updateNodeParams.Channels = numChannels
				updateNodeParams.Peers = numPeers

				i.NodeResolver.Repository.UpdateNode(ctx, updateNodeParams)
			} else {
				// Create node
				createNodeParams := db.CreateNodeParams{
					Pubkey:     getInfoResponse.IdentityPubkey,
					Addr:       addr,
					Alias:      getInfoResponse.Alias,
					Color:      getInfoResponse.Color,
					CommitHash: getInfoResponse.CommitHash,
					Version:    getInfoResponse.Version,
					Channels:   numChannels,
					Peers:      numPeers,
				}

				i.NodeResolver.Repository.CreateNode(ctx, createNodeParams)
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
