package ito

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
)

func NewCdrSessionIto(cdr db.Cdr) *SessionIto {
	return &SessionIto{
		Uid:              cdr.Uid,
		StartDatetime:    cdr.StartDateTime,
		EndDatetime:      util.NilTime(cdr.StopDateTime.Time),
		Currency:         cdr.Currency,
		TotalCost:        util.NilFloat64(cdr.TotalCost),
		TotalTime:        util.NilFloat64(cdr.TotalTime),
		TotalParkingTime: util.NilFloat64(cdr.TotalParkingTime),
		TotalSessionTime: util.NilFloat64(util.DefaultFloat(cdr.TotalTime, 0) + util.DefaultFloat(cdr.TotalParkingTime, 0)),
		TotalEnergy:      cdr.TotalEnergy,
		IsCdr:            true,
		LastUpdated:      util.DefaultTime(cdr.StopDateTime, cdr.LastUpdated),
	}
}
