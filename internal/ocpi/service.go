package ocpi

import (
	"os"

	"github.com/satimoto/go-lsp/internal/util"
	"github.com/satimoto/go-ocpi-api/ocpirpc/commandrpc"
	"github.com/satimoto/go-ocpi-api/ocpirpc/tokenrpc"
	"google.golang.org/grpc"
)

type Ocpi interface {
	GetCommandClient() commandrpc.CommandServiceClient
	GetTokenClient() tokenrpc.TokenServiceClient
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

func (s *OcpiService) GetCommandClient() commandrpc.CommandServiceClient {
	if s.commandClient == nil {
		client := commandrpc.NewCommandServiceClient(s.clientConn)
		s.commandClient = &client
	}

	return *s.commandClient
}

func (s *OcpiService) GetTokenClient() tokenrpc.TokenServiceClient {
	if s.commandClient == nil {
		client := tokenrpc.NewTokenServiceClient(s.clientConn)
		s.tokenClient = &client
	}

	return *s.tokenClient
}
