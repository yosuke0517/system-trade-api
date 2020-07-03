package service

import (
	api "app/api/bitflyer"
	"app/infrastructure/databases"
	"time"
)

// キャンドル情報を保存する
func CreateCandleWithDuration(ticker api.Ticker, productCode string, duration time.Duration) bool {
	currentCandle := databases.Select(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	// 秒単位は毎回insertして欲しい
	if currentCandle == nil {
		candle := databases.NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		candle.Insert()
		return true
	}
	// 分・時単位は秒単位ではupdateする
	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	currentCandle.Volume += ticker.Volume
	currentCandle.Close = price
	currentCandle.Save()
	return false
}
