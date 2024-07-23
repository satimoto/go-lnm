package cdr_test

import (
	"context"
	"sync"
	"time"

	"github.com/appleboy/go-fcm"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ferp/pkg/rate"
	cdrMocks "github.com/satimoto/go-lnm/internal/cdr/mocks"
	ferpMocks "github.com/satimoto/go-lnm/internal/ferp/mocks"
	lightningnetworkMocks "github.com/satimoto/go-lnm/internal/lightningnetwork/mocks"
	invoiceMocks "github.com/satimoto/go-lnm/internal/monitor/invoice/mocks"
	notificationMocks "github.com/satimoto/go-lnm/internal/notification/mocks"
	serviceMocks "github.com/satimoto/go-lnm/internal/service/mocks"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestProcessCdrErrors(t *testing.T) {
	cases := []struct {
		desc   string
		before func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockNotificationService *notificationMocks.MockNotificationService, mockOcpiService *ocpiMocks.MockOcpiService)
		cdr    db.Cdr
		after  func(*testing.T, *dbMocks.MockRepositoryService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		err    *string
	}{{
		desc: "Missing AuthorizationID",
		cdr: db.Cdr{
			Uid: "CDR0001",
		},
		err: util.NilString("cdr authorization ID is nil"),
	}, {
		desc: "Missing CDR session",
		cdr: db.Cdr{
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("error retrieving cdr session"),
	}, {
		desc: "Missing session user",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockNotificationService *notificationMocks.MockNotificationService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
			}})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("error retrieving session user"),
	}, {
		desc: "Missing session location",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockNotificationService *notificationMocks.MockNotificationService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
			}})

			mockRepository.SetGetUserMockData(dbMocks.UserMockData{User: db.User{
				CommissionPercent: 7,
			}})
		},
		cdr: db.Cdr{
			ID:              1,
			Uid:             "CDR0001",
			AuthorizationID: util.SqlNullString("AUTH0001"),
		},
		err: util.NilString("error retrieving session location"),
	}, {
		desc: "Missing session token authorization",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockNotificationService *notificationMocks.MockNotificationService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
			}})

			mockRepository.SetGetUserMockData(dbMocks.UserMockData{User: db.User{
				CommissionPercent: 7,
			}})

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
		err: nil,
	}, {
		desc: "Success",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockNotificationService *notificationMocks.MockNotificationService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetSessionByAuthorizationIDMockData(dbMocks.SessionMockData{Session: db.Session{
				ID:              1,
				Uid:             "SESSION0001",
				AuthorizationID: util.SqlNullString("AUTH0001"),
				Currency:        "EUR",
				LocationID:      2,
			}})

			mockRepository.SetGetUserMockData(dbMocks.UserMockData{User: db.User{
				CommissionPercent: 7,
			}})

			mockRepository.SetGetLocationMockData(dbMocks.LocationMockData{Location: db.Location{
				ID:      2,
				Country: "DEU",
			}})

			mockRepository.SetListSessionInvoicesBySessionIDMockData(dbMocks.SessionInvoicesMockData{SessionInvoices: []db.SessionInvoice{{
				PriceFiat:      0.3852,
				CommissionFiat: 0.026964,
				TaxFiat:        0.07831116,
				TotalFiat:      0.49047516,
				Currency:       "EUR",
			}}})

			mockFerpService.SetGetRateMockData(&rate.CurrencyRate{
				Rate:        4500,
				RateMsat:    4500000,
				LastUpdated: *util.ParseTime("2015-03-16T10:10:02Z", nil),
			})

			mockRepository.SetGetTokenAuthorizationByAuthorizationIDMockData(dbMocks.TokenAuthorizationMockData{TokenAuthorization: db.TokenAuthorization{}})

			mockNotificationService.SetSendNotificationMockData(&fcm.Response{})
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

			if sessionInvoice.PriceFiat != 0.6148 {
				t.Errorf("Error price mismatch: %v expecting %v", sessionInvoice.PriceFiat, 0.6148)
			}

			if sessionInvoice.CommissionFiat != 0.043036 {
				t.Errorf("Error commission mismatch: %v expecting %v", sessionInvoice.CommissionFiat, 0.043036)
			}

			if sessionInvoice.TaxFiat != 0.12498884 {
				t.Errorf("Error tax mismatch: %v expecting %v", sessionInvoice.TaxFiat, 0.12498884)
			}

			if sessionInvoice.TotalFiat != 0.78282484 {
				t.Errorf("Error total mismatch: %v expecting %v", sessionInvoice.TotalFiat, 0.78282484)
			}
		},
		err: nil,
	}}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mockRepository := dbMocks.NewMockRepositoryService()
			mockFerpService := ferpMocks.NewService()
			mockLightningService := lightningnetworkMocks.NewService()
			mockNotificationService := notificationMocks.NewService()
			mockOcpiService := ocpiMocks.NewService()
			mockServices := serviceMocks.NewService(mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
			cdrResolver := cdrMocks.NewResolver(mockRepository, mockServices)

			if tc.before != nil {
				tc.before(mockRepository, mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
			}

			err := cdrResolver.ProcessCdr(tc.cdr)

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
	t.Run("No session invoice", func(t *testing.T) {
		mockRepository := dbMocks.NewMockRepositoryService()
		mockFerpService := ferpMocks.NewService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockNotificationService := notificationMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		mockServices := serviceMocks.NewService(mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
		cdrResolver := cdrMocks.NewResolver(mockRepository, mockServices)

		cdr := db.Cdr{
			Uid: "CDR0001",
		}

		err := cdrResolver.ProcessCdr(cdr)

		if err.Error() == "Cdr AuthorizationID is nil" {
			t.Errorf("Error mismatch: %v expecting %v", err.Error(), "Cdr AuthorizationID is nil")
		}

	})

	t.Run("Session invoice settled", func(t *testing.T) {
		shutdownCtx, cancelFunc := context.WithCancel(context.Background())
		waitGroup := &sync.WaitGroup{}

		mockRepository := dbMocks.NewMockRepositoryService()
		mockFerpService := ferpMocks.NewService()
		mockLightningService := lightningnetworkMocks.NewService()
		mockOcpiService := ocpiMocks.NewService()
		mockNotificationService := notificationMocks.NewService()
		mockServices := serviceMocks.NewService(mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
		invoiceMonitor := invoiceMocks.NewInvoiceMonitor(mockRepository, mockServices)
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
