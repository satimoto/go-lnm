package lsp

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
)

func (s *LspService) UpdateInvoiceRequest(ctx context.Context, in *lsprpc.UpdateInvoiceRequestRequest, opts ...grpc.CallOption) (*lsprpc.UpdateInvoiceRequestResponse, error) {
	timerStart := time.Now()
	response, err := s.getInvoiceClient().UpdateInvoiceRequest(ctx, in, opts...)
	timerStop := time.Now()

	log.Printf("UpdateInvoiceRequest responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LspService) UpdateSessionInvoice(ctx context.Context, in *lsprpc.UpdateSessionInvoiceRequest, opts ...grpc.CallOption) (*lsprpc.UpdateSessionInvoiceResponse, error) {
	timerStart := time.Now()
	response, err := s.getInvoiceClient().UpdateSessionInvoice(ctx, in, opts...)
	timerStop := time.Now()

	log.Printf("UpdateSessionInvoice responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LspService) getInvoiceClient() lsprpc.InvoiceServiceClient {
	if s.invoiceClient == nil {
		client := lsprpc.NewInvoiceServiceClient(s.clientConn)
		s.invoiceClient = &client
	}

	return *s.invoiceClient
}
