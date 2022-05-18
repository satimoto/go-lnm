package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/channelrequest"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *channelrequest.ChannelRequestResolver {
	repo := channelrequest.ChannelRequestRepository(repositoryService)

	return &channelrequest.ChannelRequestResolver{
		Repository: repo,
	}
}
