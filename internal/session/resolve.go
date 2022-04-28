package session

import (
	"context"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/countryaccount"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/location"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/ocpi"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-lsp/internal/user"
)

type SessionRepository interface {
	CreateSessionInvoice(ctx context.Context, arg db.CreateSessionInvoiceParams) (db.SessionInvoice, error)
	GetSessionByUid(ctx context.Context, uid string) (db.Session, error)
	GetSessionInvoiceByPaymentRequest(ctx context.Context, paymentRequest string) (db.SessionInvoice, error)
	ListChargingPeriodDimensions(ctx context.Context, chargingPeriodID int64) ([]db.ChargingPeriodDimension, error)
	ListSessionChargingPeriods(ctx context.Context, sessionID int64) ([]db.ChargingPeriod, error)
	ListSessionInvoices(ctx context.Context, sessionID int64) ([]db.SessionInvoice, error)
	UpdateSessionInvoice(ctx context.Context, arg db.UpdateSessionInvoiceParams) (db.SessionInvoice, error)
}

type SessionResolver struct {
	Repository             SessionRepository
	LightningService       lightningnetwork.LightningNetwork
	NotificationService    notification.Notification
	OcpiService            ocpi.Ocpi
	CountryAccountResolver *countryaccount.CountryAccountResolver
	LocationResolver       *location.LocationResolver
	TariffResolver         *tariff.TariffResolver
	UserResolver           *user.UserResolver
}

func NewResolver(repositoryService *db.RepositoryService) *SessionResolver {
	lightningService := lightningnetwork.NewService()
	repo := SessionRepository(repositoryService)

	return &SessionResolver{
		Repository:             repo,
		LightningService:       lightningService,
		OcpiService:            ocpi.NewService(),
		NotificationService:    notification.NewService(),
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationResolver:       location.NewResolver(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolver(repositoryService),
	}
}
