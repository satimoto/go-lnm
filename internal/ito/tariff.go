package ito

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
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
	Elements                 []*ElementIto         `json:"elements"`
	Restriction              *TariffRestrictionIto `json:"restriction,omitempty"`
	IsIntermediateCdrCapable bool                  `json:"isIntermediateCdrCapable"`
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
