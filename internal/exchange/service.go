package exchange

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/satimoto/go-lsp/internal/exchange/kraken"
)

type Exchange interface {
	Start(ctx context.Context, waitGroup *sync.WaitGroup)
	GetRate(currency string) (*CurrencyRate, error)
}

type ExchangeService struct {
	krakenClient kraken.Kraken
	quit         bool
}

func NewService() Exchange {
	return &ExchangeService{
		krakenClient: kraken.NewExchange(),
		quit:         false,
	}
}

func (s *ExchangeService) Start(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting Exchange service")
	waitGroup.Add(1)

	go s.updateRates()

	go func() {
		<-ctx.Done()
		log.Printf("Shutting down Exchange service")

		s.stopService()

		log.Printf("Exchange service shut down")
		waitGroup.Done()
	}()
}

func (s *ExchangeService) GetRate(currency string) (*CurrencyRate, error) {
	currencyRate, err := s.krakenClient.GetRate(currency)

	if err != nil {
		return nil, err
	}

	return &CurrencyRate{
		LastUpdated: currencyRate.LastUpdated,
		Rate:        currencyRate.Rate,
		RateMsat:    currencyRate.RateMsat,
	}, nil
}

func (s *ExchangeService) updateRates() {
	for {
		s.krakenClient.UpdateRates()

		time.Sleep(1 * time.Minute)

		if s.quit {
			break
		}
	}
}

func (s *ExchangeService) stopService() {
	s.quit = true
}
