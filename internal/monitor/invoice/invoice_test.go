package invoice_test

import (
	//"encoding/json"

	"context"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	dbMocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-datastore/db"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	invoiceMocks "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	ocpiMocks "github.com/satimoto/go-lsp/internal/ocpi/mocks"

	"testing"
)

func TestInvoice(t *testing.T) {

	t.Run("No session invoice", func(t *testing.T) {
		shutdownCtx, cancelFunc := context.WithCancel(context.Background())
		waitGroup := &sync.WaitGroup{}

		mockRepository := dbMocks.NewMockRepositoryService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockLightningService, mockOcpiService)
		recvChan := mockLightningService.NewSubscribeInvoicesMockData()

		invoiceMonitor.StartMonitor(shutdownCtx, waitGroup)

		recvChan <- &lnrpc.Invoice{
			PaymentRequest: "TestPaymentRequest",
			Settled:        true,
		}

		time.Sleep(time.Second * 2)

		cancelFunc()
		waitGroup.Wait()
	})

	t.Run("Session invoice settled", func(t *testing.T) {
		shutdownCtx, cancelFunc := context.WithCancel(context.Background())
		waitGroup := &sync.WaitGroup{}

		mockRepository := dbMocks.NewMockRepositoryService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockLightningService, mockOcpiService)
		recvChan := mockLightningService.NewSubscribeInvoicesMockData()

		invoiceMonitor.StartMonitor(shutdownCtx, waitGroup)

		mockRepository.SetGetSessionInvoiceByPaymentRequestMockData(dbMocks.SessionInvoiceMockData{
			SessionInvoice: db.SessionInvoice{
				PaymentRequest: "TestPaymentRequest",
				Settled:        false,
			},
		})

		recvChan <- &lnrpc.Invoice{
			PaymentRequest: "TestPaymentRequest",
			Settled:        true,
		}

		time.Sleep(time.Second * 2)

		sessionInvoice, err := mockRepository.GetUpdateSessionInvoiceMockData()

		if err != nil {
			t.Error(err)
		}

		if sessionInvoice.Settled != true {
			t.Error("Session not settled")
		}

		cancelFunc()
		waitGroup.Wait()
	})
}
