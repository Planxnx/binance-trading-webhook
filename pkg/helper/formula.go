package helper

import (
	"github.com/shopspring/decimal"
)

func MovingUpTargetCaulculate(changePercentage float64, leverage int, entryPrice decimal.Decimal) decimal.Decimal {
	movingupPercentage := (1 + (changePercentage / float64(leverage)))
	return entryPrice.Mul(decimal.NewFromFloat(movingupPercentage)).Round(MinimuPricePrecision)
}

func MovingDownTargetCaulculate(changePercentage float64, leverage int, entryPrice decimal.Decimal) decimal.Decimal {
	movingdownPercentage := (1 - (changePercentage / float64(leverage)))
	return entryPrice.Mul(decimal.NewFromFloat(movingdownPercentage)).Round(MinimuPricePrecision)
}

func OrderQuantityCalculate(leverage int, usdSize decimal.Decimal, price decimal.Decimal) decimal.Decimal {
	// Quantity = (Leverage x USD_Size) / ETH_Price
	return decimal.NewFromInt(int64(leverage)).Mul(usdSize).DivRound(price, MinimuQtyPrecision)
}
