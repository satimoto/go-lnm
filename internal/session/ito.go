package session

import (
	"context"
	"time"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-datastore/util"
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
	StartDatetime   time.Time            `json:"start_datetime"`
	EndDatetime     *time.Time           `json:"end_datetime,omitempty"`
	Kwh             float64              `json:"kwh"`
	Currency        string               `json:"currency"`
	ChargingPeriods []*ChargingPeriodIto `json:"charging_periods"`
	TotalCost       *float64             `json:"total_cost,omitempty"`
	Status          db.SessionStatusType `json:"status"`
	LastUpdated     time.Time            `json:"last_updated"`
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
		StartDatetime: session.StartDatetime,
		EndDatetime:   util.NilTime(session.EndDatetime.Time),
		Kwh:           session.Kwh,
		Currency:      session.Currency,
		TotalCost:     util.NilFloat64(session.TotalCost.Float64),
		Status:        session.Status,
		LastUpdated:   session.LastUpdated,
	}
}

func (r *SessionResolver) CreateChargingPeriodIto(ctx context.Context, chargingPeriod db.ChargingPeriod) *ChargingPeriodIto {
	chargingPeriodIto := NewChargingPeriodIto(chargingPeriod)

	if chargingPeriodDimensions, err := r.Repository.ListChargingPeriodDimensions(ctx, chargingPeriod.ID); err == nil {
		chargingPeriodIto.Dimensions = r.CreateChargingPeriodDimensionListIto(ctx, chargingPeriodDimensions)
	}

	return chargingPeriodIto
}

func (r *SessionResolver) CreateChargingPeriodListIto(ctx context.Context, chargingPeriods []db.ChargingPeriod) []*ChargingPeriodIto {
	list := []*ChargingPeriodIto{}
	for _, chargingPeriod := range chargingPeriods {
		list = append(list, r.CreateChargingPeriodIto(ctx, chargingPeriod))
	}
	return list
}

func (r *SessionResolver) CreateChargingPeriodDimensionIto(ctx context.Context, chargingPeriodDimension db.ChargingPeriodDimension) *ChargingPeriodDimensionIto {
	return NewChargingPeriodDimensionIto(chargingPeriodDimension)
}

func (r *SessionResolver) CreateChargingPeriodDimensionListIto(ctx context.Context, chargingPeriodDimensions []db.ChargingPeriodDimension) []*ChargingPeriodDimensionIto {
	list := []*ChargingPeriodDimensionIto{}
	for _, chargingPeriodDimension := range chargingPeriodDimensions {
		list = append(list, r.CreateChargingPeriodDimensionIto(ctx, chargingPeriodDimension))
	}
	return list
}

func (r *SessionResolver) CreateSessionIto(ctx context.Context, session db.Session) *SessionIto {
	sessionIto := NewSessionIto(session)

	if chargingPeriods, err := r.Repository.ListSessionChargingPeriods(ctx, session.ID); err == nil {
		sessionIto.ChargingPeriods = r.CreateChargingPeriodListIto(ctx, chargingPeriods)
	}

	return sessionIto
}
