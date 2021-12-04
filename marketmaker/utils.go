package marketmaker

import (
	"github.com/adshao/go-binance/v2/futures"
)

func ReverseSide(side futures.SideType) futures.SideType {
	if side == futures.SideTypeBuy {
		return futures.SideTypeSell
	}
	return futures.SideTypeBuy
}
