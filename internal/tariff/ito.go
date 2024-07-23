package tariff

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lnm/internal/ito"
)

func (r *TariffResolver) CreateElementIto(ctx context.Context, element db.Element) *ito.ElementIto {
	elementIto := &ito.ElementIto{}

	if priceComponents, err := r.Repository.ListPriceComponents(ctx, element.ID); err == nil {
		elementIto.PriceComponents = r.CreatePriceComponentListIto(ctx, priceComponents)
	}

	if element.ElementRestrictionID.Valid {
		if restriction, err := r.Repository.GetElementRestriction(ctx, element.ElementRestrictionID.Int64); err == nil {
			elementIto.Restrictions = r.CreateElementRestrictionIto(ctx, restriction)
		}
	}

	return elementIto
}

func (r *TariffResolver) CreateElementListIto(ctx context.Context, elements []db.Element) []*ito.ElementIto {
	list := []*ito.ElementIto{}
	for _, element := range elements {
		list = append(list, r.CreateElementIto(ctx, element))
	}
	return list
}

func (r *TariffResolver) CreateElementRestrictionIto(ctx context.Context, elementRestriction db.ElementRestriction) *ito.ElementRestrictionIto {
	response := ito.NewElementRestrictionIto(elementRestriction)

	if weekdays, err := r.Repository.ListElementRestrictionWeekdays(ctx, elementRestriction.ID); err == nil && len(weekdays) > 0 {
		response.DayOfWeek = r.CreateWeekdayListIto(ctx, weekdays)
	}

	return response
}

func (r *TariffResolver) CreatePriceComponentIto(ctx context.Context, priceComponent db.PriceComponent) *ito.PriceComponentIto {
	priceComponentIto := ito.NewPriceComponentIto(priceComponent)

	if priceComponent.PriceRoundingID.Valid {
		if priceComponentRounding, err := r.Repository.GetPriceComponentRounding(ctx, priceComponent.PriceRoundingID.Int64); err == nil {
			priceComponentIto.PriceRound = r.CreatePriceComponentRoundingIto(ctx, priceComponentRounding)
		}
	}

	if priceComponent.StepRoundingID.Valid {
		if priceComponentRounding, err := r.Repository.GetPriceComponentRounding(ctx, priceComponent.StepRoundingID.Int64); err == nil {
			priceComponentIto.StepRound = r.CreatePriceComponentRoundingIto(ctx, priceComponentRounding)
		}
	}

	return priceComponentIto
}

func (r *TariffResolver) CreatePriceComponentListIto(ctx context.Context, priceComponents []db.PriceComponent) []*ito.PriceComponentIto {
	list := []*ito.PriceComponentIto{}
	for _, priceComponent := range priceComponents {
		list = append(list, r.CreatePriceComponentIto(ctx, priceComponent))
	}
	return list
}

func (r *TariffResolver) CreatePriceComponentRoundingIto(ctx context.Context, priceComponentRounding db.PriceComponentRounding) *ito.PriceComponentRoundingIto {
	return ito.NewPriceComponentRoundingIto(priceComponentRounding)
}

func (r *TariffResolver) CreateTariffIto(ctx context.Context, tariff db.Tariff) *ito.TariffIto {
	tariffIto := &ito.TariffIto{
		Currency: tariff.Currency,
	}

	if elements, err := r.Repository.ListElements(ctx, tariff.ID); err == nil {
		tariffIto.Elements = r.CreateElementListIto(ctx, elements)
	}

	if tariff.TariffRestrictionID.Valid {
		if tariffRestriction, err := r.Repository.GetTariffRestriction(ctx, tariff.TariffRestrictionID.Int64); err == nil {
			tariffIto.Restriction = r.CreateTariffRestrictionIto(ctx, tariffRestriction)
		}
	}

	return tariffIto
}

func (r *TariffResolver) CreateTariffRestrictionIto(ctx context.Context, tariffRestriction db.TariffRestriction) *ito.TariffRestrictionIto {
	tariffRestrictionIto := ito.NewTariffRestrictionIto(tariffRestriction)

	if weekdays, err := r.Repository.ListTariffRestrictionWeekdays(ctx, tariffRestriction.ID); err == nil && len(weekdays) > 0 {
		tariffRestrictionIto.DayOfWeek = r.CreateWeekdayListIto(ctx, weekdays)
	}

	return tariffRestrictionIto
}

func (r *TariffResolver) CreateWeekdayListIto(ctx context.Context, weekdays []db.Weekday) []*string {
	list := []*string{}
	for _, weekday := range weekdays {
		text := weekday.Text
		list = append(list, &text)
	}
	return list
}
