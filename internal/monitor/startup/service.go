package startup

import (
	"context"
	"log"
	"sync"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/cdr"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/internal/session"
)

type Startup interface {
	Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup)
}

type StartupService struct {
	CdrResolver     *cdr.CdrResolver
	SessionResolver *session.SessionResolver
	shutdownCtx     context.Context
	waitGroup       *sync.WaitGroup
	nodeID          int64
}

func NewService(repositoryService *db.RepositoryService, services *service.ServiceResolver) Startup {
	return &StartupService{
		CdrResolver:     cdr.NewResolver(repositoryService, services),
		SessionResolver: session.NewResolver(repositoryService, services),
	}
}

func (s *StartupService) Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Running Startup service")
	s.nodeID = nodeID
	s.shutdownCtx = shutdownCtx
	s.waitGroup = waitGroup

	go s.handleStartup()
}

func (s *StartupService) handleStartup() {
	s.CdrResolver.Startup(s.nodeID)
	s.SessionResolver.Startup(s.nodeID)
}
