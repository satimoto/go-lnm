package invoicerequest

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/invoicerequest"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/service"
)

type InvoiceRequestResolver struct {
	Repository          invoicerequest.InvoiceRequestRepository
	NotificationService notification.Notification
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *InvoiceRequestResolver {
	return &InvoiceRequestResolver{
		Repository:          invoicerequest.NewRepository(repositoryService),
		NotificationService: services.NotificationService,
	}
}
