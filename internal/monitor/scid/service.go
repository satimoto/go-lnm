package scid

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/pkg/util"
)

type Scid interface {
	Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup)
	AllocateScid(ctx context.Context) (*db.NodeScid, error)
}

type ScidService struct {
	LightningService lightningnetwork.LightningNetwork
	NodeRepository   node.NodeRepository
	Mutex            *sync.Mutex
	shutdownCtx      context.Context
	waitGroup        *sync.WaitGroup
	nodeID           int64
}

func NewService(repositoryService *db.RepositoryService, services *service.ServiceResolver) Scid {
	return &ScidService{
		LightningService: services.LightningService,
		NodeRepository:   node.NewRepository(repositoryService),
		Mutex:            &sync.Mutex{},
	}
}

func (s *ScidService) Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Scid service")
	s.nodeID = nodeID
	s.shutdownCtx = shutdownCtx
	s.waitGroup = waitGroup

	go s.allocateScids()
}

func (s *ScidService) AllocateScid(ctx context.Context) (*db.NodeScid, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	nodeScid, err := s.NodeRepository.GetNodeScid(ctx, s.nodeID)

	if err != nil {
		metrics.RecordError("LSP162", "Error getting node scid", err)
		log.Printf("LSP162: NodeID=%v", s.nodeID)
		return nil, errors.New("Error getting node scid")
	}

	err = s.NodeRepository.DeleteNodeScid(ctx, nodeScid.ID)

	if err != nil {
		metrics.RecordError("LSP163", "Error deleting node scid", err)
		log.Printf("LSP153: NodeScidID=%v", nodeScid.ID)
		return nil, errors.New("Error deleting node scid")
	}

	go s.allocateScid()

	return &nodeScid, nil
}

func (s *ScidService) allocateScid() {
	ctx := context.Background()
	alias, err := s.LightningService.AllocateAlias(&lnrpc.AllocateAliasRequest{})

	if err != nil {
		metrics.RecordError("LSP164", "Error allocating alias", err)
		return
	}

	shortChanID := lnwire.NewShortChanIDFromInt(alias.Scid)
	log.Printf("Allocating alias scid: %v", shortChanID.String())
	scidBytes := util.Uint64ToBytes(alias.Scid)

	createNodeScidParams := db.CreateNodeScidParams{
		NodeID: s.nodeID,
		Scid:   scidBytes,
	}

	_, err = s.NodeRepository.CreateNodeScid(ctx, createNodeScidParams)

	if err != nil {
		metrics.RecordError("LSP165", "Error creating node scid", err)
		log.Printf("LSP165: Params=%#v", createNodeScidParams)
	}
}

func (s *ScidService) allocateScids() {
	ctx := context.Background()
	scidCacheSize := dbUtil.GetEnvInt32("SCID_CACHE_SIZE", 10)

	scidCount, err := s.NodeRepository.CountNodeScids(ctx, s.nodeID)

	if err != nil {
		metrics.RecordError("LSP166", "Error getting node scid count", err)
		log.Printf("LSP166: NodeID=%v", s.nodeID)
		return
	}

	if int32(scidCount) < scidCacheSize {
		for i := int32(scidCount); i <= scidCacheSize; i++ {
			s.allocateScid()
		}
	}
}
