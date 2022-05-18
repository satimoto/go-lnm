package session

import (
	"context"
	"os"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/countryaccount"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/location"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-lsp/internal/user"
	"github.com/satimoto/go-ocpi-api/pkg/ocpi"
)

type SessionRepository interface {
	CreateSessionInvoice(ctx context.Context, arg db.CreateSessionInvoiceParams) (db.SessionInvoice, error)
	GetSessionByAuthorizationID(ctx context.Context, authorizationID string) (db.Session, error)
	GetSessionByUid(ctx context.Context, uid string) (db.Session, error)
	GetSessionInvoiceByPaymentRequest(ctx context.Context, paymentRequest string) (db.SessionInvoice, error)
	ListChargingPeriodDimensions(ctx context.Context, chargingPeriodID int64) ([]db.ChargingPeriodDimension, error)
	ListSessionChargingPeriods(ctx context.Context, sessionID int64) ([]db.ChargingPeriod, error)
	ListSessionInvoices(ctx context.Context, sessionID int64) ([]db.SessionInvoice, error)
	ListUnsettledSessionInvoicesByUserID(ctx context.Context, userID int64) ([]db.SessionInvoice, error)
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
	notificationService := notification.NewService()
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return NewResolverWithServices(repositoryService, lightningService, notificationService, ocpiService)
}

func NewResolverWithServices(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork, notificationService notification.Notification, ocpiService ocpi.Ocpi) *SessionResolver {
	repo := SessionRepository(repositoryService)

	return &SessionResolver{
		Repository:             repo,
		LightningService:       lightningService,
		OcpiService:            ocpiService,
		NotificationService:    notificationService,
		CountryAccountResolver: countryaccount.NewResolver(repositoryService),
		LocationResolver:       location.NewResolver(repositoryService),
		TariffResolver:         tariff.NewResolver(repositoryService),
		UserResolver:           user.NewResolverWithServices(repositoryService, ocpiService),
	}
}
