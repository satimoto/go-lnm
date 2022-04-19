package transaction

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionMonitor struct {
	*grpc.ClientConn
	lnrpc.LightningClient
	MacaroonCtx        context.Context
	TransactionsClient lnrpc.Lightning_SubscribeTransactionsClient
	*channelrequest.ChannelRequestResolver
}

func NewTransactionMonitor(repositoryService *db.RepositoryService) *TransactionMonitor {
	return &TransactionMonitor{
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}

func (m *TransactionMonitor) SetClientConnection(clientConn *grpc.ClientConn, macaroonCtx context.Context) {
	m.ClientConn = clientConn
	m.LightningClient = lnrpc.NewLightningClient(clientConn)
	m.MacaroonCtx = macaroonCtx
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
	htlcEventsClient, err := m.waitForSubscribeTransactionsClient(m.MacaroonCtx, 0, 1000)
	util.PanicOnError("Error creating Transactions client", err)
	m.TransactionsClient = htlcEventsClient

	for {
		htlcInterceptRequest, err := m.TransactionsClient.Recv()

		if err == nil {
			transactionChan <- *htlcInterceptRequest
		} else {
			m.TransactionsClient, err = m.waitForSubscribeTransactionsClient(m.MacaroonCtx, 100, 1000)
			util.PanicOnError("Error creating Transactions client", err)
		}
	}
}

func (m *TransactionMonitor) waitForTransactions(ctx context.Context, waitGroup *sync.WaitGroup, transactionChan chan lnrpc.Transaction) {
	waitGroup.Add(1)
	defer close(transactionChan)
	defer m.ClientConn.Close()
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

func (m *TransactionMonitor) waitForSubscribeTransactionsClient(ctx context.Context, initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeTransactionsClient, err := m.LightningClient.SubscribeTransactions(ctx, &lnrpc.GetTransactionsRequest{})

		if err == nil {
			return subscribeTransactionsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Transactions client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
