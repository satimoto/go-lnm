package exchange

import "time"

type CurrencyRate struct {
	LastUpdated time.Time
	Rate        int64
	RateMsat    int64
}