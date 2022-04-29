package session_test

import (
	//"encoding/json"

	"github.com/lightningnetwork/lnd/lnrpc"
	//dbMocks "github.com/satimoto/go-datastore-mocks/db"
	lightningnetworkMocks "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	//sessionMocks "github.com/satimoto/go-lsp/internal/session/mocks"

	"testing"
)

func TestIssueInvoice(t *testing.T) {

	t.Run("Empty", func(t *testing.T) {
		//mockRepository := dbMocks.NewMockRepositoryService()
		mockLightningService := lightningnetworkMocks.NewService()
		//sessionResolver := sessionMocks.NewResolver(mockRepository, mockLightningService)

		mockLightningService.SetAddInvoiceMockData(&lnrpc.AddInvoiceResponse{
			PaymentRequest: "mmmm",
		})

	})
}