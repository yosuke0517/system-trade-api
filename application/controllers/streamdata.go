package controllers

import (
	api "app/api/bitflyer"
	"app/config"
	"app/domain/service"
	"log"
	"os"
)

func StreamIngestionData() {
	var tickerChannl = make(chan api.Ticker)
	apiClient := api.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	go apiClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannl)
	for ticker := range tickerChannl {
		log.Printf("action=StreamIngestionData, %v", ticker)
		for _, duration := range config.Config.Durations {
			isCreated := service.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
			if isCreated == true && duration == config.Config.TradeDuration {
				// TODO
			}
		}
	}
}
