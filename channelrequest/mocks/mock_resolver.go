package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-lsp/channelrequest"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *channelrequest.ChannelRequestResolver {
	repo := channelrequest.ChannelRequestRepository(repositoryService)

	return &channelrequest.ChannelRequestResolver{
		Repository: repo,
	}
}
