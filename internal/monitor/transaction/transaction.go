package transaction

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-datastore/util"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionMonitor struct {
	LightningService       lightningnetwork.LightningNetwork
	TransactionsClient     lnrpc.Lightning_SubscribeTransactionsClient
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
}

func NewTransactionMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *TransactionMonitor {
	return &TransactionMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}

func (m *TransactionMonitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Transactions")
	transactionChan := make(chan lnrpc.Transaction)

	go m.waitForTransactions(ctx, waitGroup, transactionChan)
	go m.subscribeTransactionInterceptions(transactionChan)
}

func (m *TransactionMonitor) handleTransaction(transaction lnrpc.Transaction) {
	/** Transaction received.
	 *
	 */

	log.Printf("Transaction: %v", transaction.TxHash)
	log.Printf("BlockHash: %v", transaction.BlockHash)
	log.Printf("BlockHeight: %v", transaction.BlockHeight)
	log.Printf("Confirmations: %v", transaction.NumConfirmations)
	log.Printf("Amount: %v", transaction.Amount)
	log.Printf("TotalFees: %v", transaction.TotalFees)
}

func (m *TransactionMonitor) subscribeTransactionInterceptions(transactionChan chan<- lnrpc.Transaction) {
	htlcEventsClient, err := m.waitForSubscribeTransactionsClient(0, 1000)
	util.PanicOnError("LSP022", "Error creating Transactions client", err)
	m.TransactionsClient = htlcEventsClient

	for {
		htlcInterceptRequest, err := m.TransactionsClient.Recv()

		if err == nil {
			transactionChan <- *htlcInterceptRequest
		} else {
			m.TransactionsClient, err = m.waitForSubscribeTransactionsClient(100, 1000)
			util.PanicOnError("LSP023", "Error creating Transactions client", err)
		}
	}
}

func (m *TransactionMonitor) waitForTransactions(ctx context.Context, waitGroup *sync.WaitGroup, transactionChan chan lnrpc.Transaction) {
	waitGroup.Add(1)
	defer close(transactionChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Transactions")
			return
		case htlcInterceptRequest := <-transactionChan:
			m.handleTransaction(htlcInterceptRequest)
		}
	}
}

func (m *TransactionMonitor) waitForSubscribeTransactionsClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeTransactionsClient, err := m.LightningService.SubscribeTransactions(&lnrpc.GetTransactionsRequest{})

		if err == nil {
			return subscribeTransactionsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Transactions client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
