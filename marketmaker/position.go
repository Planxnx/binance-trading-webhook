package marketmaker

import (
	"binance-trading-webhook/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (m *MarketMaker) WaitForAvailblePosition(ctx context.Context, symbol string) error {
	getAccountService := m.client.NewGetAccountService()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if account, err := getAccountService.Do(ctx); err != nil {
				return errors.Wrap(err, "get account data")
			} else {
				for _, position := range account.Positions {

					if position.Symbol != symbol {
						continue
					}

					positionAmt, err := decimal.NewFromString(position.PositionAmt)
					if err != nil {
						return errors.Wrap(err, "convert position amount string to decimal")
					}

					if positionAmt.Sign() == 0 {
						return nil
					}

					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	}
}

func (m *MarketMaker) CloseAllPosition(ctx context.Context) error {
	getAccountService := m.client.NewGetAccountService()
	orderService := m.client.NewCreateOrderService().Symbol(m.symbol)

	if account, err := getAccountService.Do(ctx); err != nil {
		return errors.Wrap(err, "get account data")
	} else {
		for _, position := range account.Positions {

			if position.Symbol != m.symbol {
				continue
			}

			positionAmt, err := decimal.NewFromString(position.PositionAmt)
			if err != nil {
				return errors.Wrap(err, "convert position amount string to decimal")
			}

			var orderSide futures.SideType
			switch positionAmt.Sign() {
			case 1:
				orderSide = futures.SideTypeBuy
			case -1:
				orderSide = futures.SideTypeSell
			case 0:
				continue
			default:
				continue
			}

			positionQty := positionAmt.Abs().String()
			orderResp, err := orderService.Type(futures.OrderTypeMarket).Side(ReverseSide(orderSide)).Quantity(positionQty).Do(ctx)
			if err != nil {
				return errors.Wrap(err, "create close-position order")
			}
			logger.Info(
				fmt.Sprintf("=== Position has been closed ===\n[%v] %v for amount  %v ", orderResp.Side, orderResp.Symbol, positionQty),
				zap.String("event", "close_position"),
				zap.String("symbol", orderResp.Symbol),
				zap.Any("side", orderResp.Side),
				zap.String("amount", positionQty),
				zap.Int("leverage", leverage),
			)
		}
	}
	return nil
}
