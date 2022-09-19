package lsp

import (
	"context"

	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
)

func (s *LspService) UpdateInvoice(ctx context.Context, in *lsprpc.UpdateInvoiceRequest, opts ...grpc.CallOption) (*lsprpc.UpdateInvoiceResponse, error) {
	return s.getInvoiceClient().UpdateInvoice(ctx, in, opts...)
}

func (s *LspService) getInvoiceClient() lsprpc.InvoiceServiceClient {
	if s.invoiceClient == nil {
		client := lsprpc.NewInvoiceServiceClient(s.clientConn)
		s.invoiceClient = &client
	}

	return *s.invoiceClient
}
