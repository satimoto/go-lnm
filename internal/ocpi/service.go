package ocpi

import (
	"context"
	"os"

	"github.com/satimoto/go-lsp/internal/util"
	"github.com/satimoto/go-ocpi-api/ocpirpc/commandrpc"
	"github.com/satimoto/go-ocpi-api/ocpirpc/tokenrpc"
	"google.golang.org/grpc"
)

type Ocpi interface {
	StopSession(ctx context.Context, in *commandrpc.StopSessionRequest, opts ...grpc.CallOption) (*commandrpc.StopSessionResponse, error)
	UpdateTokens(ctx context.Context, in *tokenrpc.UpdateTokensRequest, opts ...grpc.CallOption) (*tokenrpc.UpdateTokensResponse, error)
}

type OcpiService struct {
	clientConn    *grpc.ClientConn
	commandClient *commandrpc.CommandServiceClient
	tokenClient   *tokenrpc.TokenServiceClient
}

func NewService() Ocpi {
	clientConn, err := grpc.Dial(os.Getenv("OCPI_RPC_ADDRESS"), grpc.WithInsecure())
	util.PanicOnError("LSP034", "Error connecting to OCPI RPC address", err)

	return &OcpiService{
		clientConn: clientConn,
	}
}

func (s *OcpiService) StopSession(ctx context.Context, in *commandrpc.StopSessionRequest, opts ...grpc.CallOption) (*commandrpc.StopSessionResponse, error) {
	return s.getCommandClient().StopSession(ctx, in, opts...)
}

func (s *OcpiService) UpdateTokens(ctx context.Context, in *tokenrpc.UpdateTokensRequest, opts ...grpc.CallOption) (*tokenrpc.UpdateTokensResponse, error) {
	return s.getTokenClient().UpdateTokens(ctx, in, opts...)
}

func (s *OcpiService) getCommandClient() commandrpc.CommandServiceClient {
	if s.commandClient == nil {
		client := commandrpc.NewCommandServiceClient(s.clientConn)
		s.commandClient = &client
	}

	return *s.commandClient
}

func (s *OcpiService) getTokenClient() tokenrpc.TokenServiceClient {
	if s.commandClient == nil {
		client := tokenrpc.NewTokenServiceClient(s.clientConn)
		s.tokenClient = &client
	}

	return *s.tokenClient
}
