package invoice_test

import (
	//"encoding/json"

	"context"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	ferpMocks "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	invoiceMocks "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	notificationMocks "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestInvoice(t *testing.T) {

	t.Run("No session invoice", func(t *testing.T) {
		shutdownCtx, cancelFunc := context.WithCancel(context.Background())
		waitGroup := &sync.WaitGroup{}

		mockRepository := dbMocks.NewMockRepositoryService()
		mockFerpService := ferpMocks.NewService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockNotificationService := notificationMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
		recvChan := mockLightningService.NewSubscribeInvoicesMockData()

		invoiceMonitor.StartMonitor(1, shutdownCtx, waitGroup)

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
		mockFerpService := ferpMocks.NewService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		mockNotificationService := notificationMocks.NewService()
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
		recvChan := mockLightningService.NewSubscribeInvoicesMockData()

		invoiceMonitor.StartMonitor(1, shutdownCtx, waitGroup)

		mockRepository.SetGetSessionInvoiceByPaymentRequestMockData(dbMocks.SessionInvoiceMockData{
			SessionInvoice: db.SessionInvoice{
				PaymentRequest: "TestPaymentRequest",
				IsSettled:      false,
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

		if sessionInvoice.IsSettled != true {
			t.Error("Session not settled")
		}

		cancelFunc()
		waitGroup.Wait()
	})
}