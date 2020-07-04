package controllers

import (
	api "app/api/bitflyer"
	"app/config"
	"app/domain/service"
	"github.com/labstack/echo"
	"github.com/valyala/fasthttp"
	"os"
)

func StreamIngestionData() {
	var tickerChannl = make(chan api.Ticker)
	apiClient := api.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	go apiClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannl)
	go func() {
		for ticker := range tickerChannl {
			for _, duration := range config.Config.Durations {
				isCreated := service.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
				if isCreated == true && duration == config.Config.TradeDuration {
					// TODO
				}
			}
		}
	}()
}

func GetAllCandle() echo.HandlerFunc {
	return func(context echo.Context) error {
		limit := 100     // TODO 動的に
		duration := "1m" // TODO 動的に
		durationTime := config.Config.Durations[duration]
		df, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), durationTime, limit)
		return context.JSON(fasthttp.StatusOK, df.Candles)
	}
}
