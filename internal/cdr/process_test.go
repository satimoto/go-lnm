package cdr_test

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-datastore/pkg/util"
	cdrMocks "github.com/satimoto/go-lsp/internal/cdr/mocks"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	invoiceMocks "github.com/satimoto/go-lsp/internal/monitor/invoice/mocks"
	notificationMocks "github.com/satimoto/go-lsp/internal/notification/mocks"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestProcessCdrErrors(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		desc   string
		before func(*dbMocks.MockRepositoryService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		cdr    db.Cdr
		after  func(*testing.T, *dbMocks.MockRepositoryService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		err    *string
	}{{
		desc: "Missing AuthorizationID",
		cdr: db.Cdr{
			Uid: "CDR0001",
		},
		err: util.NilString("Cdr AuthorizationID is nil"),
	}, {
		desc: "Missing CDR session",
		cdr: db.Cdr{
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("Error retrieving cdr session"),
	}, {
		desc: "Missing session invoices",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
			}})

			mockRepository.SetListSessionInvoicesMockData(dbMocks.SessionInvoicesMockData{Error: errors.New("Database error")})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("Error retrieving session invoices"),
	}, {
		desc: "Missing session location",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
			}})

			mockRepository.SetListSessionInvoicesMockData(dbMocks.SessionInvoicesMockData{SessionInvoices: []db.SessionInvoice{{
				AmountFiat:     0.3852,
				CommissionFiat: 0.021,
				TaxFiat:        0.0642,
				Currency:       "EUR",
			}}})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("Error retrieving session location"),
	}, {
		desc: "Missing session user",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
				LocationID:      2,
			}})

			mockRepository.SetListSessionInvoicesMockData(dbMocks.SessionInvoicesMockData{SessionInvoices: []db.SessionInvoice{{
				AmountFiat:     0.3852,
				CommissionFiat: 0.021,
				TaxFiat:        0.0642,
				Currency:       "EUR",
			}}})

			mockRepository.SetGetLocationMockData(dbMocks.LocationMockData{Location: db.Location{
				ID:      2,
				Country: "DEU",
			}})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("Error retrieving session user"),
	}, {
		desc: "Missing session user",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
				Currency:        "EUR",
				LocationID:      2,
			}})

			mockRepository.SetListSessionInvoicesMockData(dbMocks.SessionInvoicesMockData{SessionInvoices: []db.SessionInvoice{{
				AmountFiat:     0.3852,
				CommissionFiat: 0.021,
				TaxFiat:        0.0642,
				Currency:       "EUR",
			}}})

			mockRepository.SetGetLocationMockData(dbMocks.LocationMockData{Location: db.Location{
				ID:      2,
				Country: "DEU",
			}})

			mockRepository.SetGetUserMockData(dbMocks.UserMockData{User: db.User{
				CommissionPercent: 7,
			}})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
			TotalCost:       1.00,
		},
		after: func(t *testing.T, mrs *dbMocks.MockRepositoryService, mlns *lightningnetworkMocks.MockLightningNetworkService, mos *ocpiMocks.MockOcpiService) {
			sessionInvoice, err := mrs.GetCreateSessionInvoiceMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if sessionInvoice.AmountFiat != 1.1540592 {
				t.Errorf("Error mismatch: %v expecting %v", sessionInvoice.AmountFiat, 1.1540592)
			}

			if sessionInvoice.CommissionFiat != 0.06291600000000001 {
				t.Errorf("Error mismatch: %v expecting %v", sessionInvoice.CommissionFiat, 0.06291600000000001)
			}

			if sessionInvoice.TaxFiat != 0.1923432 {
				t.Errorf("Error mismatch: %v expecting %v", sessionInvoice.TaxFiat, 0.1923432)
			}
		},
		err: nil,
	}}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mockRepository := dbMocks.NewMockRepositoryService()
			mockLightningService := lightningnetworkMocks.NewService()
			mockOcpiService := ocpiMocks.NewService()
			cdrResolver := cdrMocks.NewResolver(mockRepository, mockLightningService, mockOcpiService)

			if tc.before != nil {
				tc.before(mockRepository, mockLightningService, mockOcpiService)
			}

			err := cdrResolver.ProcessCdr(ctx, tc.cdr)

			if tc.after != nil {
				tc.after(t, mockRepository, mockLightningService, mockOcpiService)
			}

			if (err == nil && tc.err != nil) || (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != *tc.err) {
				t.Errorf("Error mismatch: '%v' expecting '%v'", err, *tc.err)
			}
		})
	}
}

func TestProcessCdr(t *testing.T) {
	ctx := context.Background()

	t.Run("No session invoice", func(t *testing.T) {
		mockRepository := dbMocks.NewMockRepositoryService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		cdrResolver := cdrMocks.NewResolver(mockRepository, mockLightningService, mockOcpiService)

		cdr := db.Cdr{
			Uid: "CDR0001",
		}

		err := cdrResolver.ProcessCdr(ctx, cdr)

		if err.Error() == "Cdr AuthorizationID is nil" {
			t.Errorf("Error mismatch: %v expecting %v", err.Error(), "Cdr AuthorizationID is nil")
		}

	})

	t.Run("Session invoice settled", func(t *testing.T) {
		shutdownCtx, cancelFunc := context.WithCancel(context.Background())
		waitGroup := &sync.WaitGroup{}

		mockRepository := dbMocks.NewMockRepositoryService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		mockNotificationService := notificationMocks.NewService()
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockLightningService, mockNotificationService, mockOcpiService)
		recvChan := mockLightningService.NewSubscribeInvoicesMockData()

		invoiceMonitor.StartMonitor(shutdownCtx, waitGroup)

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
