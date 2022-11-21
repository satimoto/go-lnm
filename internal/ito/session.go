package ito

import (
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
)

type ChargingPeriodIto struct {
	StartDateTime time.Time                     `json:"start_date_time"`
	Dimensions    []*ChargingPeriodDimensionIto `json:"dimensions"`
}

type ChargingPeriodDimensionIto struct {
	Type   db.ChargingPeriodDimensionType `json:"type"`
	Volume float64                        `json:"volume"`
}

type SessionIto struct {
	Uid              string               `json:"uid"`
	StartDatetime    time.Time            `json:"start_datetime"`
	EndDatetime      *time.Time           `json:"end_datetime,omitempty"`
	Currency         string               `json:"currency"`
	TotalCost        *float64             `json:"total_cost,omitempty"`
	TotalTime        *float64             `json:"total_time,omitempty"`
	TotalParkingTime *float64             `json:"total_parking_time,omitempty"`
	TotalSessionTime *float64             `json:"total_session_time,omitempty"`
	TotalEnergy      float64              `json:"total_energy"`
	ChargingPeriods  []*ChargingPeriodIto `json:"charging_periods"`
	LastUpdated      time.Time            `json:"last_updated"`
}

func NewChargingPeriodIto(chargingPeriod db.ChargingPeriod) *ChargingPeriodIto {
	return &ChargingPeriodIto{
		StartDateTime: chargingPeriod.StartDateTime,
	}
}

func NewChargingPeriodDimensionIto(chargingPeriodDimension db.ChargingPeriodDimension) *ChargingPeriodDimensionIto {
	return &ChargingPeriodDimensionIto{
		Type:   chargingPeriodDimension.Type,
		Volume: chargingPeriodDimension.Volume,
	}
}

func NewSessionIto(session db.Session) *SessionIto {
	return &SessionIto{
		Uid:           session.Uid,
		StartDatetime: session.StartDatetime,
		EndDatetime:   util.NilTime(session.EndDatetime.Time),
		Currency:      session.Currency,
		TotalCost:     util.NilFloat64(session.TotalCost.Float64),
		TotalEnergy:   session.Kwh,
		LastUpdated:   session.LastUpdated,
	}
}
