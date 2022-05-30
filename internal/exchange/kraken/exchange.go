package kraken

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/satimoto/go-datastore/pkg/util"
)

const (
	krakenAPIURL  = "https://api.kraken.com"
	krakenTicker  = "Ticker"
	krakenVersion = "0"
)

type Kraken interface {
	UpdateRates()
	GetRate(currency string) (*CurrencyRate, error)
}

type KrakenExchange struct {
	httpClient    *http.Client
	currencyRates LatestCurrencyRates
}

func NewExchange() Kraken {
	return &KrakenExchange{
		httpClient:    http.DefaultClient,
		currencyRates: make(LatestCurrencyRates),
	}
}

func NewExchangeWithClient(httpClient *http.Client) Kraken {
	return &KrakenExchange{
		httpClient:    httpClient,
		currencyRates: make(LatestCurrencyRates),
	}
}

func (e *KrakenExchange) UpdateRates() {
	values := url.Values{}
	values.Set("pair", "XBTEUR,XBTGBP")
	values.Set("interval", "1")

	requestUrl := fmt.Sprintf("%s/%s/public/%s?%s", krakenAPIURL, krakenVersion, krakenTicker, values.Encode())
	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)

	if err != nil {
		util.LogOnError("LSP050", "Error forming request", err)
		log.Printf("LSP050: Url=%v", requestUrl)
		return
	}

	response, err := e.httpClient.Do(request)

	if err != nil {
		util.LogOnError("LSP051", "Error making request", err)
		util.LogHttpRequest("LSP051", requestUrl, request, false)
		return
	}

	tickerResponse, err := UnmarshalTickerResponse(response.Body)

	if err != nil {
		util.LogOnError("LSP052", "Error unmarshalling response", err)
		util.LogHttpResponse("LSP051", requestUrl, response, false)
		return
	}

	for pair, value := range tickerResponse.Data {
		currency := getCurrency(pair)
		price, err := strconv.ParseFloat(value.Last[0], 64)

		if err != nil {
			util.LogOnError("LSP053", "Error parsing float", err)
			log.Printf("LSP050: Value=%v", value.Last[0])
			continue
		}

		e.currencyRates[currency] = CurrencyRate{
			LastUpdated: time.Now(),
			Rate:        int64(100_000_000 / price),
			RateMsat:    int64(100_000_000_000 / price),
		}

		log.Printf("%s: %v sats / %v millisats", currency, e.currencyRates[currency].Rate, e.currencyRates[currency].RateMsat)
	}
}

func (e *KrakenExchange) GetRate(currency string) (*CurrencyRate, error) {
	if currencyRate, ok := e.currencyRates[currency]; ok {
		return &currencyRate, nil
	}

	return nil, errors.New("no currency rate available")
}

func getCurrency(pair string) string {
	return string(pair[len(pair)-3:])
}
