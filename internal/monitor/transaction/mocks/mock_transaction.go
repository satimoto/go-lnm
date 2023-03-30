package mocks

import (
	mocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-lnm/internal/monitor/transaction"
	"github.com/satimoto/go-lnm/internal/service"
)

func NewTransactionMonitor(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *transaction.TransactionMonitor {
	return &transaction.TransactionMonitor{
		LightningService:       services.LightningService,
	}
}
