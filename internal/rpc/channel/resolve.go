package channel

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
)

type RpcChannelResolver struct {
	LightningService       lightningnetwork.LightningNetwork
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *RpcChannelResolver {
	return &RpcChannelResolver{
		LightningService:       services.LightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
