package lsp

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Lsp interface {
	OpenChannel(ctx context.Context, in *lsprpc.OpenChannelRequest, opts ...grpc.CallOption) (*lsprpc.OpenChannelResponse, error)
	UpdateInvoice(ctx context.Context, in *lsprpc.UpdateInvoiceRequest, opts ...grpc.CallOption) (*lsprpc.UpdateInvoiceResponse, error)
}

type LspService struct {
	clientConn    *grpc.ClientConn
	channelClient *lsprpc.ChannelServiceClient
	invoiceClient *lsprpc.InvoiceServiceClient
}

func NewService(address string) Lsp {
	clientConn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	util.PanicOnError("LSP108", "Error connecting to LSP RPC address", err)

	return &LspService{
		clientConn: clientConn,
	}
}
