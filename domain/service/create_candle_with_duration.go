package service

import (
	"app/api/bitflyer"
	"app/domain/model"
	"app/infrastructure/databases/candle"
	"time"
)

// キャンドル情報を保存する
func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	currentCandle := candle.SelectOne(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	// 秒単位は毎回insert
	if currentCandle == nil {
		candle := candle.NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
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

// chart?product_code=FX_BTC_JPY&duration=1h
func GetAllCandle(productCode string, duration time.Duration, limit int) (dfCandle *model.DataFrameCandle, err error) {
	rows := candle.SelectAll(productCode, duration, limit)
	defer rows.Close()
	dfCandle = &model.DataFrameCandle{}
	dfCandle.ProductCode = productCode
	dfCandle.Duration = duration
	for rows.Next() {
		var candle model.Candle
		candle.ProductCode = productCode
		candle.Duration = duration
		rows.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return dfCandle, nil
}
