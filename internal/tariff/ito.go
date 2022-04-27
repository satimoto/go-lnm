package tariff

import (
	"context"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-datastore/util"
)

type ElementIto struct {
	PriceComponents []*PriceComponentIto   `json:"price_components"`
	Restrictions    *ElementRestrictionIto `json:"restrictions,omitempty"`
}

type ElementRestrictionIto struct {
	StartTime   *string   `json:"start_time,omitempty"`
	EndTime     *string   `json:"end_time,omitempty"`
	StartDate   *string   `json:"start_date,omitempty"`
	EndDate     *string   `json:"end_date,omitempty"`
	MinKwh      *float64  `json:"min_kwh,omitempty"`
	MaxKwh      *float64  `json:"max_kwh,omitempty"`
	MinPower    *float64  `json:"min_power,omitempty"`
	MaxPower    *float64  `json:"max_power,omitempty"`
	MinDuration *int32    `json:"min_duration,omitempty"`
	MaxDuration *int32    `json:"max_duration,omitempty"`
	DayOfWeek   []*string `json:"day_of_week"`
}

type PriceComponentIto struct {
	Type                db.TariffDimension         `json:"type"`
	Price               float64                    `json:"price"`
	StepSize            int32                      `json:"step_size"`
	PriceRound          *PriceComponentRoundingIto `json:"price_round,omitempty"`
	StepRound           *PriceComponentRoundingIto `json:"step_round,omitempty"`
	ExactPriceComponent *bool                      `json:"exact_price_component"`
}

type PriceComponentRoundingIto struct {
	Granularity db.RoundingGranularity `json:"granularity"`
	Rule        db.RoundingRule        `json:"rule"`
}

type TariffIto struct {
	Elements    []*ElementIto         `json:"elements"`
	Restriction *TariffRestrictionIto `json:"restriction,omitempty"`
}

type TariffRestrictionIto struct {
	StartTime  *string   `json:"start_time,omitempty"`
	EndTime    *string   `json:"end_time,omitempty"`
	StartTime2 *string   `json:"start_time_2,omitempty"`
	EndTime2   *string   `json:"end_time_2,omitempty"`
	DayOfWeek  []*string `json:"day_of_week"`
}

func NewElementRestrictionIto(elementRestriction db.ElementRestriction) *ElementRestrictionIto {
	return &ElementRestrictionIto{
		StartTime:   util.NilString(elementRestriction.StartTime),
		EndTime:     util.NilString(elementRestriction.EndTime),
		StartDate:   util.NilString(elementRestriction.StartDate),
		EndDate:     util.NilString(elementRestriction.EndDate),
		MinKwh:      util.NilFloat64(elementRestriction.MinKwh.Float64),
		MaxKwh:      util.NilFloat64(elementRestriction.MaxKwh.Float64),
		MinPower:    util.NilFloat64(elementRestriction.MinPower.Float64),
		MaxPower:    util.NilFloat64(elementRestriction.MaxPower.Float64),
		MinDuration: util.NilInt32(elementRestriction.MinDuration.Int32),
		MaxDuration: util.NilInt32(elementRestriction.MaxDuration.Int32),
	}
}

func NewPriceComponentIto(priceComponent db.PriceComponent) *PriceComponentIto {
	return &PriceComponentIto{
		Type:                priceComponent.Type,
		Price:               priceComponent.Price,
		StepSize:            priceComponent.StepSize,
		ExactPriceComponent: util.NilBool(priceComponent.ExactPriceComponent),
	}
}

func NewPriceComponentRoundingIto(priceComponentRounding db.PriceComponentRounding) *PriceComponentRoundingIto {
	return &PriceComponentRoundingIto{
		Granularity: priceComponentRounding.Granularity,
		Rule:        priceComponentRounding.Rule,
	}
}

func NewTariffRestrictionIto(tariffRestriction db.TariffRestriction) *TariffRestrictionIto {
	return &TariffRestrictionIto{
		StartTime:  &tariffRestriction.StartTime,
		EndTime:    &tariffRestriction.EndTime,
		StartTime2: util.NilString(tariffRestriction.StartTime2),
		EndTime2:   util.NilString(tariffRestriction.EndTime2),
	}
}

func (r *TariffResolver) CreateElementIto(ctx context.Context, element db.Element) *ElementIto {
	elementIto := &ElementIto{}

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

func (r *TariffResolver) CreateElementListIto(ctx context.Context, elements []db.Element) []*ElementIto {
	list := []*ElementIto{}
	for _, element := range elements {
		list = append(list, r.CreateElementIto(ctx, element))
	}
	return list
}

func (r *TariffResolver) CreateElementRestrictionIto(ctx context.Context, elementRestriction db.ElementRestriction) *ElementRestrictionIto {
	response := NewElementRestrictionIto(elementRestriction)

	if weekdays, err := r.Repository.ListElementRestrictionWeekdays(ctx, elementRestriction.ID); err == nil && len(weekdays) > 0 {
		response.DayOfWeek = r.CreateWeekdayListIto(ctx, weekdays)
	}

	return response
}

func (r *TariffResolver) CreatePriceComponentIto(ctx context.Context, priceComponent db.PriceComponent) *PriceComponentIto {
	priceComponentIto := NewPriceComponentIto(priceComponent)

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

func (r *TariffResolver) CreatePriceComponentListIto(ctx context.Context, priceComponents []db.PriceComponent) []*PriceComponentIto {
	list := []*PriceComponentIto{}
	for _, priceComponent := range priceComponents {
		list = append(list, r.CreatePriceComponentIto(ctx, priceComponent))
	}
	return list
}

func (r *TariffResolver) CreatePriceComponentRoundingIto(ctx context.Context, priceComponentRounding db.PriceComponentRounding) *PriceComponentRoundingIto {
	return NewPriceComponentRoundingIto(priceComponentRounding)
}

func (r *TariffResolver) CreateTariffIto(ctx context.Context, tariff db.Tariff) *TariffIto {
	tariffIto := &TariffIto{}

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

func (r *TariffResolver) CreateTariffRestrictionIto(ctx context.Context, tariffRestriction db.TariffRestriction) *TariffRestrictionIto {
	tariffRestrictionIto := NewTariffRestrictionIto(tariffRestriction)

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
