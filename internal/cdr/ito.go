package cdr

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/ito"
)

func (r *CdrResolver) CreateChargingPeriodIto(ctx context.Context, chargingPeriod db.ChargingPeriod) *ito.ChargingPeriodIto {
	chargingPeriodIto := ito.NewChargingPeriodIto(chargingPeriod)

	if chargingPeriodDimensions, err := r.SessionResolver.Repository.ListChargingPeriodDimensions(ctx, chargingPeriod.ID); err == nil {
		chargingPeriodIto.Dimensions = r.CreateChargingPeriodDimensionListIto(ctx, chargingPeriodDimensions)
	}

	return chargingPeriodIto
}

func (r *CdrResolver) CreateChargingPeriodListIto(ctx context.Context, chargingPeriods []db.ChargingPeriod) []*ito.ChargingPeriodIto {
	list := []*ito.ChargingPeriodIto{}
	for _, chargingPeriod := range chargingPeriods {
		list = append(list, r.CreateChargingPeriodIto(ctx, chargingPeriod))
	}
	return list
}

func (r *CdrResolver) CreateChargingPeriodDimensionIto(ctx context.Context, chargingPeriodDimension db.ChargingPeriodDimension) *ito.ChargingPeriodDimensionIto {
	return ito.NewChargingPeriodDimensionIto(chargingPeriodDimension)
}

func (r *CdrResolver) CreateChargingPeriodDimensionListIto(ctx context.Context, chargingPeriodDimensions []db.ChargingPeriodDimension) []*ito.ChargingPeriodDimensionIto {
	list := []*ito.ChargingPeriodDimensionIto{}
	for _, chargingPeriodDimension := range chargingPeriodDimensions {
		list = append(list, r.CreateChargingPeriodDimensionIto(ctx, chargingPeriodDimension))
	}
	return list
}

func (r *CdrResolver) CreateSessionIto(ctx context.Context, cdr db.Cdr) *ito.SessionIto {
	sessionIto := ito.NewCdrSessionIto(cdr)

	if chargingPeriods, err := r.Repository.ListCdrChargingPeriods(ctx, cdr.ID); err == nil {
		sessionIto.ChargingPeriods = r.CreateChargingPeriodListIto(ctx, chargingPeriods)
	}

	return sessionIto
}
