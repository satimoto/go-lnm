package cdr_test

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	cdrMocks "github.com/satimoto/go-lsp/internal/cdr/mocks"
	ferpMocks "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestIssueInvoiceRequest(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		desc        string
		before      func(*dbMocks.MockRepositoryService, *ferpMocks.MockFerpService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		amountMsat int64
		after       func(*testing.T, *dbMocks.MockRepositoryService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		err         *string
	}{{
		desc:        "200 millisats",
		amountMsat: 200,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 200 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 200)
			}
		},
	}, {
		desc:        "1000 millisats",
		amountMsat: 1000,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 1000 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 1000)
			}
		},
	}, {
		desc:        "100000 millisats",
		amountMsat: 100000,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 100000 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 100000)
			}
		},
	}, {
		desc:        "50000 millisats",
		amountMsat: 50000,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 50000 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 50000)
			}
		},
	}, {
		desc:        "11174911 millisats",
		amountMsat: 11174911,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 11174911 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 11174911)
			}
		},
	}, {
		desc: "1000000 millisats",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetUnsettledInvoiceRequestByPromotionCodeMockData(dbMocks.InvoiceRequestMockData{InvoiceRequest: db.InvoiceRequest{
				ID:          3,
				UserID:      1,
				PromotionID: 2,
				AmountMsat:  1000,
			}})
		},
		amountMsat: 5000,
		err:         nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetUpdateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.AmountMsat != 6000 {
				t.Errorf("Error mismatch: %v expecting %v", invoiceRequest.AmountMsat, 6000)
			}
		},
	}}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mockRepository := dbMocks.NewMockRepositoryService()
			mockFerpService := ferpMocks.NewService()
			mockLightningService := lightningnetworkMocks.NewService()
			mockOcpiService := ocpiMocks.NewService()
			cdrResolver := cdrMocks.NewResolver(mockRepository, mockFerpService, mockLightningService, mockOcpiService)

			if tc.before != nil {
				tc.before(mockRepository, mockFerpService, mockLightningService, mockOcpiService)
			}

			mockRepository.SetGetUserMockData(dbMocks.UserMockData{User: db.User{
				ID: 2,
			}})

			mockRepository.SetGetPromotionByCodeMockData(dbMocks.PromotionMockData{Promotion: db.Promotion{
				Code: "CIRCUIT",
			}})

			err := cdrResolver.IssueInvoiceRequest(ctx, 1, "CIRCUIT", tc.amountMsat)

			if tc.after != nil {
				tc.after(t, mockRepository, mockLightningService, mockOcpiService)
			}

			if (err == nil && tc.err != nil) || (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != *tc.err) {
				t.Errorf("Error mismatch: '%v' expecting '%v'", err, *tc.err)
			}
		})
	}
}
