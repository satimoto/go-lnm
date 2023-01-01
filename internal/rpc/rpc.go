package rpc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/monitor"
	"github.com/satimoto/go-lsp/internal/rpc/cdr"
	"github.com/satimoto/go-lsp/internal/rpc/channel"
	"github.com/satimoto/go-lsp/internal/rpc/invoice"
	"github.com/satimoto/go-lsp/internal/rpc/rpc"
	"github.com/satimoto/go-lsp/internal/rpc/session"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/lsprpc"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"google.golang.org/grpc"
)

type Rpc interface {
	StartRpc(*sync.WaitGroup)
}

type RpcService struct {
	RepositoryService  *db.RepositoryService
	Server             *grpc.Server
	RpcCdrResolver     *cdr.RpcCdrResolver
	RpcChannelResolver *channel.RpcChannelResolver
	RpcInvoiceResolver *invoice.RpcInvoiceResolver
	RpcResolver        *rpc.RpcResolver
	RpcSessionResolver *session.RpcSessionResolver
	ShutdownCtx        context.Context
}

func NewRpc(shutdownCtx context.Context, d *sql.DB, services *service.ServiceResolver, monitorService *monitor.Monitor) Rpc {
	repositoryService := db.NewRepositoryService(d)

	return &RpcService{
		RepositoryService:  repositoryService,
		Server:             grpc.NewServer(),
		RpcCdrResolver:     cdr.NewResolver(repositoryService, services),
		RpcChannelResolver: channel.NewResolver(repositoryService, services, monitorService.ScidService),
		RpcInvoiceResolver: invoice.NewResolver(repositoryService, services),
		RpcResolver:        rpc.NewResolver(repositoryService, services),
		RpcSessionResolver: session.NewResolver(repositoryService, services),
		ShutdownCtx:        shutdownCtx,
	}
}

func (rs *RpcService) StartRpc(waitGroup *sync.WaitGroup) {
	log.Printf("Starting Rpc service")
	waitGroup.Add(1)

	go rs.listenAndServe()

	go func() {
		<-rs.ShutdownCtx.Done()
		log.Printf("Shutting down Rpc service")

		rs.shutdown()

		log.Printf("Rpc service shut down")
		waitGroup.Done()
	}()
}

func (rs *RpcService) listenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("RPC_PORT")))
	util.PanicOnError("LSP028", "Error creating network address", err)

	lsprpc.RegisterChannelServiceServer(rs.Server, rs.RpcChannelResolver)
	lsprpc.RegisterInvoiceServiceServer(rs.Server, rs.RpcInvoiceResolver)
	ocpirpc.RegisterCdrServiceServer(rs.Server, rs.RpcCdrResolver)
	ocpirpc.RegisterRpcServiceServer(rs.Server, rs.RpcResolver)
	ocpirpc.RegisterSessionServiceServer(rs.Server, rs.RpcSessionResolver)

	err = rs.Server.Serve(listener)
	metrics.RecordError("LSP029", "Error in Rpc service", err)
}

func (rs *RpcService) shutdown() {
	rs.Server.GracefulStop()
}
