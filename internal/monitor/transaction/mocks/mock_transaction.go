package mocks

import (
	mocks "github.com/satimoto/go-datastore-mocks/db"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	lightningnetwork "github.com/satimoto/go-lsp/internal/lightningnetwork/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
)

func NewTransactionMonitor(repositoryService *mocks.MockRepositoryService, lightningService *lightningnetwork.MockLightningNetworkService) *transaction.TransactionMonitor {
	return &transaction.TransactionMonitor{
		LightningService:       lightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
