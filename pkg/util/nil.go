package util

import (
	"database/sql"

	dbUtil "github.com/satimoto/go-datastore/pkg/util"
)

func AddNullFloat64(floatA, floatB sql.NullFloat64) sql.NullFloat64 {
	return dbUtil.SqlNullFloat64(floatA.Float64 + floatB.Float64)
}

func AddNullInt64(intA, intB sql.NullInt64) sql.NullInt64 {
	return dbUtil.SqlNullInt64(intA.Int64 + intB.Int64)
}

func MinusNullFloat64(floatA, floatB sql.NullFloat64) sql.NullFloat64 {
	return dbUtil.SqlNullFloat64(floatA.Float64 - floatB.Float64)
}

func MinusNullInt64(intA, intB sql.NullInt64) sql.NullInt64 {
	return dbUtil.SqlNullInt64(intA.Int64 - intB.Int64)
}
