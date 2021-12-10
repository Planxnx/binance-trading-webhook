package marketmaker

import (
	"binance-trading-webhook/pkg/logger"
	"os"
	"strconv"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var (
	stopLoss      = 0.45
	takeProfit    = 1.0
	leverage      = 40
	defaultsymbol = "ETHUSDT"
)

type MarketMaker struct {
	client *futures.Client

	symbol          string
	leverage        int
	stopLoss        float64
	takeProfit      float64
	leverageDecimal decimal.Decimal
}

func init() {

	//TODO: refactor this to use a config file

	if err := godotenv.Load(); err != nil {
		logger.Error("Failed to load .env", zap.Error(err))
	}

	tp := os.Getenv("TAKE_PROFIT")
	if tp != "" {
		if s, err := strconv.ParseFloat(tp, 64); err == nil {
			takeProfit = s
		}
	}
	sl := os.Getenv("STOP_LOSS")
	if sl != "" {
		if s, err := strconv.ParseFloat(sl, 64); err == nil {
			stopLoss = s
		}
	}

	lv := os.Getenv("LEVERAGE")
	if lv != "" {
		if s, err := strconv.Atoi(lv); err == nil {
			leverage = s
		}
	}

}

func New(symbol string, client *futures.Client) *MarketMaker {

	//FIXME: this is a hack to get the symbol from the env
	if symbol == "" {
		symbol = defaultsymbol
	}
	return &MarketMaker{
		client:          client,
		symbol:          symbol,
		leverage:        leverage,
		stopLoss:        stopLoss,
		takeProfit:      takeProfit,
		leverageDecimal: decimal.NewFromInt(int64(leverage)),
	}
}
