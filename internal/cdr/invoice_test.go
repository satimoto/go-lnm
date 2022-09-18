package cdr_test

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ferp/pkg/rate"
	cdrMocks "github.com/satimoto/go-lsp/internal/cdr/mocks"
	ferpMocks "github.com/satimoto/go-lsp/internal/ferp/mocks"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/pkg/util"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestIssueInvoiceRequest(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		desc          string
		before        func(*dbMocks.MockRepositoryService, *ferpMocks.MockFerpService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		invoiceParams util.InvoiceParams
		after         func(*testing.T, *dbMocks.MockRepositoryService, *lightningnetworkMocks.MockLightningNetworkService, *ocpiMocks.MockOcpiService)
		err           *string
	}{{
		desc: "Create invoice request",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockFerpService.SetGetRateMockData(&rate.CurrencyRate{
				Rate:        4500,
				RateMsat:    4500000,
				LastUpdated: *dbUtil.ParseTime("2015-03-16T10:10:02Z", nil),
			})
		},
		invoiceParams: util.InvoiceParams{
			PriceFiat:      dbUtil.SqlNullFloat64(0.3852),
			CommissionFiat: dbUtil.SqlNullFloat64(0.026964),
			TaxFiat:        dbUtil.SqlNullFloat64(0.07831116),
			TotalFiat:      dbUtil.SqlNullFloat64(0.49047516),
		},
		err: nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.PriceFiat.Float64 != 0.3852 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceFiat.Float64, 0.3852)
			}

			if invoiceRequest.PriceMsat.Int64 != 1733400 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceMsat.Int64, 1733400)
			}

			if invoiceRequest.CommissionFiat.Float64 != 0.026964 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionFiat.Float64, 0.026964)
			}

			if invoiceRequest.CommissionMsat.Int64 != 121338 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionMsat.Int64, 121338)
			}

			if invoiceRequest.TaxFiat.Float64 != 0.07831116 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxFiat.Float64, 0.07831116)
			}

			if invoiceRequest.TaxMsat.Int64 != 352400 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxMsat.Int64, 352400)
			}

			if invoiceRequest.TotalFiat != 0.49047516 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalFiat, 0.49047516)
			}

			if invoiceRequest.TotalMsat != 2207138 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalMsat, 2207138)
			}
		},
	}, {
		desc: "Update invoice request",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockRepository.SetGetUnsettledInvoiceRequestMockData(dbMocks.InvoiceRequestMockData{InvoiceRequest: db.InvoiceRequest{
				ID:             3,
				UserID:         1,
				PromotionID:    2,
				PriceFiat:      dbUtil.SqlNullFloat64(0.3852),
				PriceMsat:      dbUtil.SqlNullInt64(1733400),
				CommissionFiat: dbUtil.SqlNullFloat64(0.026964),
				CommissionMsat: dbUtil.SqlNullInt64(121338),
				TaxFiat:        dbUtil.SqlNullFloat64(0.07831116),
				TaxMsat:        dbUtil.SqlNullInt64(352400),
				TotalFiat:      0.49047516,
				TotalMsat:      2207138,
			}})

			mockFerpService.SetGetRateMockData(&rate.CurrencyRate{
				Rate:        4500,
				RateMsat:    4500000,
				LastUpdated: *dbUtil.ParseTime("2015-03-16T10:10:02Z", nil),
			})
		},
		invoiceParams: util.InvoiceParams{
			PriceFiat:      dbUtil.SqlNullFloat64(0.3852),
			CommissionFiat: dbUtil.SqlNullFloat64(0.026964),
			TaxFiat:        dbUtil.SqlNullFloat64(0.07831116),
			TotalFiat:      dbUtil.SqlNullFloat64(0.49047516),
		},
		err: nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetUpdateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.PriceFiat.Float64 != 0.7704 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceFiat.Float64, 0.7704)
			}

			if invoiceRequest.PriceMsat.Int64 != 3466800 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceMsat.Int64, 3466800)
			}

			if invoiceRequest.CommissionFiat.Float64 != 0.053928 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionFiat.Float64, 0.053928)
			}

			if invoiceRequest.CommissionMsat.Int64 != 242676 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionMsat.Int64, 242676)
			}

			if invoiceRequest.TaxFiat.Float64 != 0.15662232 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxFiat.Float64, 0.15662232)
			}

			if invoiceRequest.TaxMsat.Int64 != 704800 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxMsat.Int64, 704800)
			}

			if invoiceRequest.TotalFiat != 0.98095032 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalFiat, 0.98095032)
			}

			if invoiceRequest.TotalMsat != 4414276 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalMsat, 4414276)
			}
		},
	}, {
		desc: "Crate invoice request",
		before: func(mockRepository *dbMocks.MockRepositoryService, mockFerpService *ferpMocks.MockFerpService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			mockFerpService.SetGetRateMockData(&rate.CurrencyRate{
				Rate:        4500,
				RateMsat:    4500000,
				LastUpdated: *dbUtil.ParseTime("2015-03-16T10:10:02Z", nil),
			})
		},
		invoiceParams: util.InvoiceParams{
			TotalMsat: dbUtil.SqlNullInt64(2207138),
		},
		err: nil,
		after: func(t *testing.T, mockRepository *dbMocks.MockRepositoryService, mockLightningService *lightningnetworkMocks.MockLightningNetworkService, mockOcpiService *ocpiMocks.MockOcpiService) {
			invoiceRequest, err := mockRepository.GetCreateInvoiceRequestMockData()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if invoiceRequest.PriceFiat.Float64 != 0 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceFiat.Float64, 0)
			}

			if invoiceRequest.PriceMsat.Int64 != 0 {
				t.Errorf("Error price mismatch: %v expecting %v", invoiceRequest.PriceMsat.Int64, 0)
			}

			if invoiceRequest.CommissionFiat.Float64 != 0 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionFiat.Float64, 0)
			}

			if invoiceRequest.CommissionMsat.Int64 != 0 {
				t.Errorf("Error commission mismatch: %v expecting %v", invoiceRequest.CommissionMsat.Int64, 0)
			}

			if invoiceRequest.TaxFiat.Float64 != 0 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxFiat.Float64, 0)
			}

			if invoiceRequest.TaxMsat.Int64 != 0 {
				t.Errorf("Error tax mismatch: %v expecting %v", invoiceRequest.TaxMsat.Int64, 0)
			}

			if invoiceRequest.TotalFiat != 0.4904751111111111 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalFiat, 0.49047516)
			}

			if invoiceRequest.TotalMsat != 2207138 {
				t.Errorf("Error total mismatch: %v expecting %v", invoiceRequest.TotalMsat, 2207138)
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

			_, err := cdrResolver.IssueInvoiceRequest(ctx, 1, "CIRCUIT", "EUR", "Satsback", tc.invoiceParams)

			if tc.after != nil {
				tc.after(t, mockRepository, mockLightningService, mockOcpiService)
			}

			if (err == nil && tc.err != nil) || (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != *tc.err) {
				t.Errorf("Error mismatch: '%v' expecting '%v'", err, *tc.err)
			}
		})
	}
}
