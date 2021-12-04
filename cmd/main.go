package main

import (
	"binance-trading-webhook/handler"
	"binance-trading-webhook/marketmaker"
	"binance-trading-webhook/marketprice"
	"binance-trading-webhook/pkg/logger"
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	loggermiddleware "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var (
	PORT               string
	Environment        string
	BINANCE_API_KEY    string
	BINANCE_API_SECRET string
	useTestNet         bool
	symbol             string
	useSize            string
	usdSizeDecimal     = decimal.Decimal{}
)

func init() {

	//TODO: refactor this to use a config file
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Failed to load .env", zap.Error(err))
	}
	symbol = os.Getenv("SYMBOL")
	BINANCE_API_KEY = os.Getenv("BINANCE_API_KEY")
	BINANCE_API_SECRET = os.Getenv("BINANCE_API_SECRET")

	USE_TEST_NET := os.Getenv("USE_TEST_NET")
	if USE_TEST_NET == "true" {
		BINANCE_API_KEY = os.Getenv("TESTNET_BINANCE_API_KEY")
		BINANCE_API_SECRET = os.Getenv("TESTNET_BINANCE_API_SECRET")
		useTestNet = true
	}

	Environment = os.Getenv("ENVIRONMENT")
	if Environment == "" {
		Environment = "development"
	}

	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	useSize = os.Getenv("USD_SIZE")
	usdSizeDecimal, err = decimal.NewFromString(useSize)
	if err != nil {
		logger.Fatal("Failed to parse USD_SIZE", zap.Error(err))
	}

}

func main() {

	//Initialize logger
	if err := logger.InitLogger(&logger.Config{
		Environment: Environment,
	}); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}

	// Initialize the Binance Futures client
	futures.UseTestnet = useTestNet
	binanceClient := futures.NewClient(BINANCE_API_KEY, BINANCE_API_SECRET)

	// Initialize Services
	mmakerClient := marketmaker.New(symbol, binanceClient)
	mpriceClient := marketprice.New(symbol)

	// Start the market price streaming service
	if err := mpriceClient.Run(); err != nil {
		logger.Fatal("Failed to start market price client", zap.Error(err))
	}

	// Initialize the Handler
	apiHandler := handler.New(mmakerClient, mpriceClient)

	// HTTP Router
	app := fiber.New(fiber.Config{
		ReadTimeout: time.Second * 5,
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(map[string]bool{"ok": true})
	})

	app.Use(cors.New())
	app.Use(loggermiddleware.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(e interface{}) {
			buf := make([]byte, 1024) // bufLen = 1024
			buf = buf[:runtime.Stack(buf, false)]
			logger.Error(fmt.Sprintf("panic: %v\n%s\n", e, buf))
		},
	}))

	tradeAPI := app.Group("/trade")
	tradeAPI.All("/close-all", apiHandler.CloseAllPositionHandler)

	webhookAPI := app.Group("/webhook")
	webhookAPI.Post("/tradingview", apiHandler.TradingViewSignalsHandler(symbol, usdSizeDecimal))

	go func() {
		err := app.Listen(fmt.Sprintf(":%s", PORT))
		if err != nil {
			logger.Fatal(err.Error(), zap.String("action", "start server"))
		}
	}()

	logger.Info("Starting server on port :" + PORT)

	// Graceful shutdown
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
		syscall.SIGTERM, // kill -SIGTERM XXXX
	)

	<-signalChan
	logger.Info("os.Interrupt - Shutting down...\n")

	shutdownTimeout := 30 * time.Second
	gracefullShutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := app.Shutdown(); err != nil {
		logger.Error(err.Error(), zap.String("action", "shutdown server"))
	} else {
		logger.Info("Gracefully stopped server")
	}

	if err := mpriceClient.Close(gracefullShutdownCtx); err != nil {
		logger.Error(err.Error(), zap.String("action", "shutdown market price client"))
	}

	if err := mmakerClient.CancelAllOpenOrdersService(gracefullShutdownCtx); err != nil {
		logger.Error(err.Error(), zap.String("action", "cancel-all-open-orders"))
	}
	if err := mmakerClient.CloseAllPosition(gracefullShutdownCtx); err != nil {
		logger.Error(err.Error(), zap.String("action", "close-all-position"))
	}

}
