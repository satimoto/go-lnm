package channel

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/channelrequest"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
)

type RpcChannelResolver struct {
	LightningService       lightningnetwork.LightningNetwork
	ChannelRequestResolver *channelrequest.ChannelRequestResolver
}

func NewResolver(repositoryService *db.RepositoryService) *RpcChannelResolver {
	lightningService := lightningnetwork.NewService()

	return NewResolverWithServices(repositoryService, lightningService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *RpcChannelResolver {
	return &RpcChannelResolver{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
