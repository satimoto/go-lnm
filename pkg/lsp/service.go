package lsp

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Lsp interface {
	OpenChannel(ctx context.Context, in *lsprpc.OpenChannelRequest, opts ...grpc.CallOption) (*lsprpc.OpenChannelResponse, error)
	ListChannels(ctx context.Context, in *lsprpc.ListChannelsRequest, opts ...grpc.CallOption) (*lsprpc.ListChannelsResponse, error)
	UpdateInvoiceRequest(ctx context.Context, in *lsprpc.UpdateInvoiceRequestRequest, opts ...grpc.CallOption) (*lsprpc.UpdateInvoiceRequestResponse, error)
	UpdateSessionInvoice(ctx context.Context, in *lsprpc.UpdateSessionInvoiceRequest, opts ...grpc.CallOption) (*lsprpc.UpdateSessionInvoiceResponse, error)
}

type LspService struct {
	clientConn    *grpc.ClientConn
	channelClient *lsprpc.ChannelServiceClient
	invoiceClient *lsprpc.InvoiceServiceClient
}

func NewService(address string) Lsp {
	timerStart := time.Now()
	clientConn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	timerStop := time.Now()

	util.PanicOnError("LSP108", "Error connecting to LSP RPC address", err)
	log.Printf("LSP %v dialed in %f seconds", address, timerStop.Sub(timerStart).Seconds())

	return &LspService{
		clientConn: clientConn,
	}
}
