package controllers

import (
	api "app/api/bitflyer"
	"app/application/response"
	"app/config"
	"app/domain/service"
	"net/http"
	"os"
	"strconv"
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

func GetAllCandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// limit := 100     // TODO 画面から設定できるように
		// duration := "1m" // TODO 画面から設定できるように
		// api/chart?product_code=FX_BTC_JPY&duration=1hとか
		productCode := r.URL.Query().Get("product_code") // TODO 後々product codeも画面から設定できるように（以下サンプル）
		if productCode == "" {
			response.BadRequest(w, "No product_code")
		}
		strLimit := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(strLimit)
		if strLimit == "" || err != nil || limit < 0 || limit > 1000 {
			limit = 1000
		}

		duration := r.URL.Query().Get("duration")
		if duration == "" {
			duration = "1m"
		}
		durationTime := config.Config.Durations[duration]

		df, _ := service.GetAllCandle(productCode, durationTime, limit)
		// df, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), durationTime, limit)
		response.Success(w, df.Candles)
	}
}
