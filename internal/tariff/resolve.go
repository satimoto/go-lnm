package tariff

import (
	"context"

	"github.com/satimoto/go-datastore/db"
)

type TariffRepository interface {
	GetElementRestriction(ctx context.Context, id int64) (db.ElementRestriction, error)
	GetPriceComponentRounding(ctx context.Context, id int64) (db.PriceComponentRounding, error)
	GetTariffByUid(ctx context.Context, uid string) (db.Tariff, error)
	GetTariffRestriction(ctx context.Context, id int64) (db.TariffRestriction, error)
	ListElements(ctx context.Context, tariffID int64) ([]db.Element, error)
	ListElementRestrictionWeekdays(ctx context.Context, elementRestrictionID int64) ([]db.Weekday, error)
	ListPriceComponents(ctx context.Context, elementID int64) ([]db.PriceComponent, error)
	ListTariffRestrictionWeekdays(ctx context.Context, tariffRestrictionID int64) ([]db.Weekday, error)
}

type TariffResolver struct {
	Repository TariffRepository
}

func NewResolver(repositoryService *db.RepositoryService) *TariffResolver {
	repo := TariffRepository(repositoryService)
	return &TariffResolver{
		Repository: repo,
	}
}
