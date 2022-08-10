package mocks

import (
	channelrequestMocks "github.com/satimoto/go-datastore/pkg/channelrequest/mocks"
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lsp/internal/channelrequest"
)

func NewResolver(repositoryService *mocks.MockRepositoryService) *channelrequest.ChannelRequestResolver {
	return &channelrequest.ChannelRequestResolver{
		Repository: channelrequestMocks.NewRepository(repositoryService),
	}
}
