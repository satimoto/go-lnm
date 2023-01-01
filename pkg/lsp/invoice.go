package lsp

import (
	"context"

	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
)

func (s *LspService) UpdateInvoiceRequest(ctx context.Context, in *lsprpc.UpdateInvoiceRequestRequest, opts ...grpc.CallOption) (*lsprpc.UpdateInvoiceRequestResponse, error) {
	return s.getInvoiceClient().UpdateInvoiceRequest(ctx, in, opts...)
}

func (s *LspService) UpdateSessionInvoice(ctx context.Context, in *lsprpc.UpdateSessionInvoiceRequest, opts ...grpc.CallOption) (*lsprpc.UpdateSessionInvoiceResponse, error) {
	return s.getInvoiceClient().UpdateSessionInvoice(ctx, in, opts...)
}

func (s *LspService) getInvoiceClient() lsprpc.InvoiceServiceClient {
	if s.invoiceClient == nil {
		client := lsprpc.NewInvoiceServiceClient(s.clientConn)
		s.invoiceClient = &client
	}

	return *s.invoiceClient
}
