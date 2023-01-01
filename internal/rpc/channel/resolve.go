package channel

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/node"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/monitor/scid"
	"github.com/satimoto/go-lsp/internal/service"
)

type RpcChannelResolver struct {
	LightningService       lightningnetwork.LightningNetwork
	ScidService            scid.Scid
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
	NodeRepository         node.NodeRepository
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver, scidService scid.Scid) *RpcChannelResolver {
	return &RpcChannelResolver{
		LightningService:       services.LightningService,
		ScidService:            scidService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
		NodeRepository:         node.NewRepository(repositoryService),
	}
}
