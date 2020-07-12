package controllers

import (
	"app/api/bitflyer"
	"app/application/response"
	"app/config"
	"app/domain/service"
	"fmt"
	"log"
	"math"
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
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
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

		// ロスカット
		if time.Now().Truncate(time.Second).Second() == 57 {
			params := map[string]string{
				"product_code":      "FX_BTC_JPY",
				"child_order_state": "ACTIVE",
			}
			orderRes, _ := bitflyerClient.ListOrder(params)
			// 注文
			if len(orderRes) != 0 {
				//orderTime, err := orderRes[0].ChildOrderDate
				//
				//if err != nil {
				//	log.Fatal("損切り時間の計算に失敗しました。bitflyerのコンソールにて損切り注文をしてください。")
				//}
				orderTime := orderRes[0].TruncateDateTime(time.Second)
				fmt.Println("orderTime")
				fmt.Println(orderTime)
				ticker, err := bitflyerClient.GetTicker(os.Getenv("PRODUCT_CODE"))
				if err != nil {
					log.Fatal("ticker情報の取得に失敗しました。アプリケーションを終了します。")
				}

				// 基準価格計算
				currentPrice := ticker.GetMidPrice()
				limitPrice := currentPrice - orderRes[0].Price
				fmt.Println("limitPrice")
				fmt.Println(math.Abs(limitPrice))
				fmt.Println("注文から30分以上経過したかどうか？")
				fmt.Println(orderTime.Add(time.Minute * 30).Before(time.Now()))
				// TODO 損切りの条件（仮）注文してから30分経過 or 注文時の価格と現在価格が3000円以上差がある時||
				if orderTime.Add(time.Minute*30).Before(time.Now()) == true || math.Abs(limitPrice) > 3000 {
					fmt.Println("損切りの条件に達したため注文をキャンセルし、成行でクローズします。")
					cancelOrder := &bitflyer.CancelOrder{
						ProductCode:            "FX_BTC_JPY",
						ChildOrderAcceptanceID: orderRes[0].ChildOrderAcceptanceID,
					}
					statusCode, _ := bitflyerClient.CancelOrder(cancelOrder)
					time.Sleep(time.Second * 1)
					if statusCode != 200 {
						log.Fatal("損切りに失敗しました。bitflyerのマイページから手動で損切りしてください。")
					}
					if statusCode == 200 {
						order := &bitflyer.Order{
							ProductCode:     "FX_BTC_JPY",
							ChildOrderType:  "MARKET",
							Side:            orderRes[0].Side,
							Size:            orderRes[0].Size,
							MinuteToExpires: 1440,
							TimeInForce:     "GTC",
						}
						closeRes, _ := bitflyerClient.SendOrder(order)
						fmt.Println("設定時間をオーバーしました。損切りします。")
						fmt.Println(closeRes)
					}
				}
			}
		}
	}

}
