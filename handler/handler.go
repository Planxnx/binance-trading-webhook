package handler

import (
	"binance-trading-webhook/marketmaker"
	"binance-trading-webhook/marketprice"
)

type Handler struct {
	mmakerClient *marketmaker.MarketMaker
	mpriceClient *marketprice.MarketPrice
}

func New(mmakerClient *marketmaker.MarketMaker, mpriceClient *marketprice.MarketPrice) *Handler {
	return &Handler{
		mmakerClient: mmakerClient,
		mpriceClient: mpriceClient,
	}
}
