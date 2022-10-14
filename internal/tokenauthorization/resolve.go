package tokenauthorization

import (
	"github.com/satimoto/go-datastore/pkg/cdr"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/tokenauthorization"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/satimoto/go-lsp/internal/session"
)

type TokenAuthorizationResolver struct {
	Repository          tokenauthorization.TokenAuthorizationRepository
	NotificationService notification.Notification
	CdrRepository       cdr.CdrRepository
	SessionResolver     session.SessionResolver
}

func NewResolver(repositoryService *db.RepositoryService, services *service.ServiceResolver) *TokenAuthorizationResolver {
	return &TokenAuthorizationResolver{
		Repository:          tokenauthorization.NewRepository(repositoryService),
		NotificationService: services.NotificationService,
		CdrRepository:       cdr.NewRepository(repositoryService),
		SessionResolver:     *session.NewResolver(repositoryService, services),
	}
}
