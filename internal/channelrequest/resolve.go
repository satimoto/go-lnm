package channelrequest

import (
	"github.com/satimoto/go-datastore/pkg/channelrequest"
	"github.com/satimoto/go-datastore/pkg/db"
)

type ChannelRequestResolver struct {
	Repository channelrequest.ChannelRequestRepository
}

func NewResolver(repositoryService *db.RepositoryService) *ChannelRequestResolver {
	return &ChannelRequestResolver{
		Repository: channelrequest.NewRepository(repositoryService),
	}
}
