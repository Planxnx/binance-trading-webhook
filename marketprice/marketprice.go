package marketprice

import (
	"context"
	"log"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type MarketPrice struct {
	symbol       string
	done         chan struct{}
	cancel       chan struct{}
	isRunning    bool
	runningMutex sync.Mutex

	//TODO: upgrade to atomic.Value or ConcurrentMap
	price        string
	priceDecimal decimal.Decimal
}

func New(symbol string) *MarketPrice {
	return &MarketPrice{
		symbol: symbol,
	}
}

// Run 	start MarketPrice Streaming Service
func (m *MarketPrice) Run() error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	marketPriceStreamDone, marketPriceStreamCancel, err := futures.WsMarkPriceServe(m.symbol, m.marketPriceHandler, m.marketPriceErrorHandler)
	if err != nil {
		return errors.Wrap(err, "start market price streaming service")
	}

	m.done = marketPriceStreamDone
	m.cancel = marketPriceStreamCancel
	m.isRunning = true

	return nil
}

//Close close market price streaming service
func (m *MarketPrice) Close(ctx context.Context) error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	if !m.isRunning {
		return nil
	}

	m.cancel <- struct{}{}

	for {
		select {
		case <-m.done:
			return nil
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "market price streaming service shutfown failed")
		}
	}
}

// IsRunning returns true if the market price streaming service is running
func (m *MarketPrice) IsRunning() bool {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	return m.isRunning
}

// GetPrice returns the current market price
func (m *MarketPrice) GetPrice() decimal.Decimal {
	return m.priceDecimal
}

// marketPriceHandler handles the market price from websocket events
func (m *MarketPrice) marketPriceHandler(event *futures.WsMarkPriceEvent) {
	m.price = event.MarkPrice
	priceDec, err := decimal.NewFromString(m.price)
	if err != nil {
		log.Printf("Market Price Handler Error: %v\n", err)
		return
	}
	m.priceDecimal = priceDec
}

// marketPriceErrorHandler handles the market price error from websocket events
func (m *MarketPrice) marketPriceErrorHandler(err error) {
	log.Printf("Market Price Streaming Error: %v\n", err)
}
