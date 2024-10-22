package transaction

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionMonitor struct {
	LightningService   lightningnetwork.LightningNetwork
	TransactionsClient lnrpc.Lightning_SubscribeTransactionsClient
	nodeID             int64
}

func NewTransactionMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *TransactionMonitor {
	return &TransactionMonitor{
		LightningService: services.LightningService,
	}
}

func (m *TransactionMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Transactions")
	transactionChan := make(chan lnrpc.Transaction)

	m.nodeID = nodeID
	go m.waitForTransactions(shutdownCtx, waitGroup, transactionChan)
	go m.subscribeTransactionInterceptions(transactionChan)
	go m.updateWalletBalance()
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

	go m.updateWalletBalance()
}

func (m *TransactionMonitor) subscribeTransactionInterceptions(transactionChan chan<- lnrpc.Transaction) {
	transactionsClient, err := m.waitForSubscribeTransactionsClient(0, 1000)
	util.PanicOnError("LNM022", "Error creating Transactions client", err)
	m.TransactionsClient = transactionsClient

	for {
		transaction, err := m.TransactionsClient.Recv()

		if err == nil {
			transactionChan <- *transaction
		} else {
			m.TransactionsClient, err = m.waitForSubscribeTransactionsClient(100, 1000)
			util.PanicOnError("LNM023", "Error creating Transactions client", err)
		}
	}
}

func (m *TransactionMonitor) updateWalletBalance() {
	walletBalance, err := m.LightningService.WalletBalance(&lnrpc.WalletBalanceRequest{})

	if err != nil {
		metrics.RecordError("LNM080", "Error requesting wallet balance", err)
	}

	log.Printf("TotalBalance: %v", walletBalance.TotalBalance)
	log.Printf("ConfirmedBalance: %v", walletBalance.ConfirmedBalance)
	log.Printf("UnconfirmedBalance: %v", walletBalance.UnconfirmedBalance)
	log.Printf("LockedBalance: %v", walletBalance.LockedBalance)

	metricWalletTotalBalanceSatoshis.Set(float64(walletBalance.TotalBalance))
	metricWalletConfirmedBalanceSatoshis.Set(float64(walletBalance.ConfirmedBalance))
	metricWalletUnconfirmedBalanceSatoshis.Set(float64(walletBalance.UnconfirmedBalance))
	metricWalletLockedBalanceSatoshis.Set(float64(walletBalance.LockedBalance))
	metricWalletReservedBalanceSatoshis.Set(float64(walletBalance.ReservedBalanceAnchorChan))
}

func (m *TransactionMonitor) waitForTransactions(shutdownCtx context.Context, waitGroup *sync.WaitGroup, transactionChan chan lnrpc.Transaction) {
	waitGroup.Add(1)
	defer close(transactionChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Transactions")
			return
		case transaction := <-transactionChan:
			m.handleTransaction(transaction)
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
