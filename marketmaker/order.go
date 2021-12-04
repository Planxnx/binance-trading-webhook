package marketmaker

import (
	"binance-trading-webhook/pkg/helper"
	"binance-trading-webhook/pkg/logger"
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (m *MarketMaker) CreateOrder(ctx context.Context, side futures.SideType, price decimal.Decimal, usdSize decimal.Decimal) error {

	positionQty := helper.OrderQuantityCalculate(m.leverage, usdSize, price)

	orderType := futures.OrderTypeMarket

	orderService := m.client.NewCreateOrderService().Symbol(m.symbol)
	leverageService := m.client.NewChangeLeverageService().Symbol(m.symbol)

	if _, err := leverageService.Leverage(m.leverage).Do(ctx); err != nil {
		return errors.Wrapf(err, "change leverage failed: %v", m.symbol)
	}

	orderResp, err := orderService.Side(side).Type(orderType).Quantity(positionQty.String()).Do(ctx)
	if err != nil {
		return errors.Wrapf(err, "create order failed:%v %v for %v %v at %v usd", side, positionQty, m.symbol, price)
	}

	logger.Info(
		fmt.Sprintf("=== Order Position has been created ===\n[%v] %v for amount %v  ", orderResp.Side, orderResp.Symbol, positionQty),
		zap.String("event", "create_order"),
		zap.String("symbol", orderResp.Symbol),
		zap.Any("side", orderResp.Side),
		zap.String("amount", positionQty.String()),
		zap.Int("leverage", leverage),
		zap.Any("order_type", orderType),
	)

	return nil

}

func (m *MarketMaker) CreateStopLoss(ctx context.Context, orderSide futures.SideType, orderPrice decimal.Decimal) error {

	var stopPrice decimal.Decimal
	switch orderSide {
	case futures.SideTypeBuy:
		stopPrice = helper.MovingDownTargetCaulculate(m.takeProfit, m.leverage, orderPrice)
	case futures.SideTypeSell:
		stopPrice = helper.MovingUpTargetCaulculate(m.takeProfit, m.leverage, orderPrice)
	}

	if err := m.CreateTPorSL(ctx, orderSide, futures.OrderTypeStopMarket, stopPrice); err != nil {
		return errors.Wrap(err, "create stoploss failed")
	}
	return nil
}

func (m *MarketMaker) CreateTakeProfit(ctx context.Context, orderSide futures.SideType, orderPrice decimal.Decimal) error {

	var stopPrice decimal.Decimal
	switch orderSide {
	case futures.SideTypeBuy:
		stopPrice = helper.MovingUpTargetCaulculate(m.takeProfit, m.leverage, orderPrice)
	case futures.SideTypeSell:
		stopPrice = helper.MovingDownTargetCaulculate(m.takeProfit, m.leverage, orderPrice)
	}

	if err := m.CreateTPorSL(ctx, orderSide, futures.OrderTypeTakeProfitMarket, stopPrice); err != nil {
		return errors.Wrap(err, "create takeprofit failed")
	}

	return nil
}

// CreateTPorSL creates take profit or stop loss order
func (m *MarketMaker) CreateTPorSL(ctx context.Context, orderSide futures.SideType, orderType futures.OrderType, stopPrice decimal.Decimal) error {

	side := ReverseSide(orderSide)

	orderService := m.client.NewCreateOrderService().Symbol(m.symbol)
	orderResp, err := orderService.Side(side).Type(orderType).StopPrice(stopPrice.String()).ClosePosition(true).Do(ctx)
	if err != nil {
		return errors.Wrapf(err, "create stoploss/takeprofiy failed: %v type, %v %v for %v at %v usd", orderType, side, m.symbol, stopPrice)
	}

	logger.Info(
		fmt.Sprintf("=== TP/SL Position has been created (for close-position)===\n[%v] %v at %v usd ", orderResp.Side, orderResp.Symbol, stopPrice),
		zap.String("event", "create_tp_sl"),
		zap.Any("order_type", orderType),
		zap.String("symbol", orderResp.Symbol),
		zap.Any("side", orderResp.Side),
		zap.String("stop_price", stopPrice.String()),
	)

	return nil
}

func (m *MarketMaker) CancelAllOpenOrdersService(ctx context.Context) error {
	orderService := m.client.NewCancelAllOpenOrdersService().Symbol(m.symbol)
	if err := orderService.Do(ctx); err != nil {
		return errors.Wrapf(err, "cancel all open orders failed: %v", m.symbol)
	}
	return nil
}
