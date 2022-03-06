package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (i *Intercept) SubscribeTransactions() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	subscribeTransactionsClient, err := i.waitForSubscribeTransactions(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating SubscribeTransactions client", err)

	for {
		if err := i.SubscribeTransaction(macaroonCtx, subscribeTransactionsClient); err != nil {
			subscribeTransactionsClient, err = i.waitForSubscribeTransactions(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating SubscribeTransactions client", err)
		}
	}
}

func (i *Intercept) SubscribeTransaction(macaroonCtx context.Context, subscribeTransactionsClient lnrpc.Lightning_SubscribeTransactionsClient) error {
	transaction, err := subscribeTransactionsClient.Recv()

	if err != nil {
		log.Printf("Error receiving tranaction event: %v", status.Code(err))

		return err
	}

	/** Transaction received.
	 *
	 */

	log.Printf("Transaction: %v", transaction.TxHash)
	log.Printf("BlockHash: %v", transaction.BlockHash)
	log.Printf("BlockHeight: %v", transaction.BlockHeight)
	log.Printf("Confirmations: %v", transaction.NumConfirmations)
	log.Printf("Amount: %v", transaction.Amount)
	log.Printf("TotalFees: %v", transaction.TotalFees)

	return nil
}

func (i *Intercept) waitForSubscribeTransactions(ctx context.Context, initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeTransactionsClient, err := i.LightningClient.SubscribeTransactions(ctx, &lnrpc.GetTransactionsRequest{})

		if err == nil {
			return subscribeTransactionsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeTransactions client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
