package helper_test

import (
	"binance-trading-webhook/pkg/helper"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMovingUpTargetFormula(t *testing.T) {

	changePercentage := 3.155 // 315.50%
	leverage := 116
	entryPrice := decimal.NewFromFloat(1_234_567.89)
	expectedTarget := decimal.NewFromFloat(1_268_146.01)

	target := helper.MovingUpTargetCaulculate(changePercentage, leverage, entryPrice)

	if expectedTarget.Cmp(target) != 0 {
		t.Errorf("Expected %v, got %v", expectedTarget, target)
	}
}

func TestMovingDownTargetFormula(t *testing.T) {

	changePercentage := 0.1 // 10%
	leverage := 3
	entryPrice := decimal.NewFromFloat(9876543.21)
	expectedTarget := decimal.NewFromFloat(9_547_325.10)

	target := helper.MovingDownTargetCaulculate(changePercentage, leverage, entryPrice)

	if expectedTarget.Cmp(target) != 0 {
		t.Errorf("Expected %v, got %v", expectedTarget, target)
	}
}

func TestOrderQuantityFormula(t *testing.T) {

	price := decimal.NewFromFloat(4_112.04)
	usdMarginSize := decimal.NewFromFloat(1.25)
	leverage := 66
	expectedQuantity := decimal.NewFromFloat(0.020)

	quantity := helper.OrderQuantityCalculate(leverage, usdMarginSize, price)

	if expectedQuantity.Cmp(quantity) != 0 {
		t.Errorf("Expected %v, got %v", expectedQuantity, quantity)
	}
}
