package controllers

import (
	"app/api/bitflyer"
	"app/application/response"
	"app/config"
	"app/domain/service"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func StreamIngestionData() {
	var tickerChannl = make(chan bitflyer.Ticker)
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	go bitflyerClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannl)
	go func() {
		for ticker := range tickerChannl {
			for _, duration := range config.Config.Durations {
				isCreated := service.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
				if isCreated == true && duration == config.Config.TradeDuration {
					fmt.Println("ticker.Timestamp")
					fmt.Println(ticker.Timestamp)
				}
			}
		}
	}()
}

// パラメータに応じた単位のローソク足情報を返す
func GetAllCandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		productCode := r.URL.Query().Get("product_code")
		if productCode == "" {
			response.BadRequest(w, "No product_code")
		}
		strLimit := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(strLimit)
		if strLimit == "" || err != nil || limit < 0 || limit > 1000 {
			// デフォルトは1000とする
			limit = 1000
		}

		duration := r.URL.Query().Get("duration")
		if duration == "" {
			duration = "1m"
		}
		durationTime := config.Config.Durations[duration]

		df, _ := service.GetAllCandle(productCode, durationTime, limit)
		response.Success(w, df.Candles)
	}
}

func SystemTradeBase() {
	//// 証拠金が設定範囲内か確認
	//collateral, err := bitflyerClient.GetCollateral()
	//i, _ := strconv.ParseFloat(os.Getenv("MIN_COLLATERAL"), 64)
	//if err != nil {
	//	log.Fatalf("action=SystemTradeBase err=%s", err.Error())
	//}
	//if collateral.Collateral < i {
	//	log.Fatal("証拠金が設定金額を下回っているため取引を中止します。")
	//}
	//var tickerChannl = make(chan bitflyer.Ticker)
	//bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	//go bitflyerClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannl)
	//go func() {
	//	for ticker := range tickerChannl {
	//		service.SystemTradeService(ticker, ticker.ProductCode)
	//	}
	//}()

	// 1秒タイマー
	for range time.Tick(1 * time.Second) {
		fmt.Println(time.Now().Truncate(time.Second))
		if time.Now().Truncate(time.Second).Second() == 0 {
			var wg sync.WaitGroup
			closeOrderExecChannel := make(chan bool)
			wg.Add(1)
			go service.CloseOrderExecutionCheck(closeOrderExecChannel, &wg)
			for c := range closeOrderExecChannel {
				fmt.Println("CloseOrderExecutionCheck 結果")
				fmt.Println(c)
				fmt.Println(time.Now().Truncate(time.Second))
				if c == true {
					go service.SystemTradeService(os.Getenv("PRODUCT_CODE"), time.Now().Truncate(time.Second))
				}
				close(closeOrderExecChannel)
			}
		}
	}

}
