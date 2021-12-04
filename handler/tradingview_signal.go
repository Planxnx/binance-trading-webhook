package handler

import (
	"net/http"
	"strings"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type tradingViewSignalsRequest struct {
	Symbol string           `json:"symbol"`
	Side   futures.SideType `json:"side"`
}

func (h *Handler) TradingViewSignalsHandler(symbol string, useSize decimal.Decimal) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		req := &tradingViewSignalsRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
		}

		if err := req.validate(); err != nil {
			return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
		}

		if req.Symbol != symbol {
			return c.Status(http.StatusBadRequest).JSON(map[string]string{
				"error": "invalid symbol, " + req.Symbol,
			})
		}

		if err := h.mmakerClient.CancelAllOpenOrdersService(c.UserContext()); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}
		if err := h.mmakerClient.CloseAllPosition(c.UserContext()); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}

		if err := h.mmakerClient.WaitForAvailblePosition(c.UserContext(), symbol); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}

		currentPrice := h.mpriceClient.GetPrice()

		if err := h.mmakerClient.CreateOrder(c.UserContext(), req.Side, currentPrice, useSize); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}

		if err := h.mmakerClient.CreateStopLoss(c.UserContext(), req.Side, currentPrice); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}

		if err := h.mmakerClient.CreateTakeProfit(c.UserContext(), req.Side, currentPrice); err != nil {
			return c.Status(500).JSON(map[string]string{"error": err.Error()})
		}

		return c.Status(http.StatusOK).JSON(map[string]string{"message": "success"})
	}
}

func (req *tradingViewSignalsRequest) validate() error {
	var validateErrors []string

	if req.Side != "BUY" && req.Side != "SELL" {
		validateErrors = append(validateErrors, "invalid side, "+string(req.Side))
	}

	if validateErrors == nil {
		return nil
	}

	return errors.Errorf("%v", strings.Join(validateErrors, ", "))
}
