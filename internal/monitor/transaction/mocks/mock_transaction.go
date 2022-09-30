package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	channelrequest "github.com/satimoto/go-lsp/internal/channelrequest/mocks"
	"github.com/satimoto/go-lsp/internal/monitor/transaction"
	"github.com/satimoto/go-lsp/internal/service"
)

func NewTransactionMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *transaction.TransactionMonitor {
	return &transaction.TransactionMonitor{
		LightningService:       services.LightningService,
		ChannelRequestResolver: channelrequest.NewResolver(repositoryService),
	}
}
