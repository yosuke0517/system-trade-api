package controllers

import (
	"app/api/bitflyer"
	"app/application/response"
	"app/config"
	"app/domain/service"
	"app/infrastructure/databases/candle"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func StreamIngestionData() {
	var menteCount = 0
	var tickerChannl = make(chan bitflyer.Ticker)
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	go bitflyerClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannl)
	go func() {
	StreamIngestionData:
		for {
			if time.Now().Truncate(time.Second).Hour() == 19 {
				if time.Now().Truncate(time.Second).Minute() < 30 {
					log.Println("4時〜4時30分までメンテナンスのため取引を中断します。")
					goto Mente
				}
			}
			for ticker := range tickerChannl {
				for _, duration := range config.Config.Durations {
					isCreated := service.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
					if isCreated == true && duration == config.Config.TradeDuration {
						fmt.Println("ticker.Timestamp")
						fmt.Println(ticker.Timestamp)
					}
				}
			}
		}
	Mente:
		for {
			for range time.Tick(1 * time.Second) {
				menteCount++
				fmt.Println(menteCount)
				if menteCount == 1800 {
					log.Println("Mente：システムトレードを再開します。")
					menteCount = 0
					goto StreamIngestionData
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
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	var isUpper = 0
	var currentCandle *service.CandleInfraStruct
	var closeOrderExecutionCheck = false
	var count = 0
	var menteCount = 0
SystemTrade:
	for {
		// 1秒タイマー
		for range time.Tick(1 * time.Second) {
			// TODO 4時台は取引しない（cronで制御？？）
			fmt.Println(time.Now().Truncate(time.Second))
			if time.Now().Truncate(time.Second).Hour() == 19 {
				if time.Now().Truncate(time.Second).Minute() < 30 {
					log.Println("4時〜4時30分までメンテナンスのため取引を中断します。")
					goto Mente
				}
			}
			// 0秒台で分析・システムトレードを走らせる
			if time.Now().Truncate(time.Second).Second() == 0 {
				if closeOrderExecutionCheck == true {
					go service.SystemTradeService(isUpper, currentCandle)
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
				if len(orderRes) == 0 {
					fmt.Println("オーダーはありません。")
				} else {
					orderTime := orderRes[0].TruncateDateTime(time.Second)
					fmt.Println("残注文の発注時間")
					fmt.Println(orderTime)
					ticker, err := bitflyerClient.GetTicker(os.Getenv("PRODUCT_CODE"))
					if err != nil {
						log.Fatal("ticker情報の取得に失敗しました。アプリケーションを終了します。")
					}

					// 基準価格計算
					currentPrice := ticker.GetMidPrice()
					limitPrice := currentPrice - orderRes[0].Price
					limitPriceAbsolute := math.Abs(limitPrice)
					fmt.Println("上限乖離値かどうか")
					fmt.Println(limitPriceAbsolute)
					log.Printf("注文した価格との乖離：%s", strconv.FormatFloat(limitPriceAbsolute, 'f', -1, 64))
					fmt.Printf("orderTime：%s", orderTime)
					fmt.Println("注文から60分以上経過したかどうか？")
					fmt.Println(orderTime.Add(time.Minute * 60).Before(time.Now()))
					// TODO 損切りの条件（仮）注文してから60分経過 or 注文時の価格と現在価格が2000円以上差がある時 ||中止中
					if orderTime.Add(time.Minute*60).Before(time.Now()) == true || math.Abs(limitPrice) > 2000 {
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
							fmt.Println("損切りorderRRRRRRRRRRRRRRRRRR")
							fmt.Println(order)
							closeRes, _ := bitflyerClient.SendOrder(order)
							log.Printf("設定時間または設定価格をオーバーしました。損切りします。%s", time.Now())
							fmt.Println(closeRes)
							if closeRes == nil {
								for i := 0; i < 5; i++ {
									closeRes, _ := bitflyerClient.SendOrder(order)
									if closeRes != nil {
										break
									}
								}
							}
						}
					}
				}
			}

			// 注文準備
			if time.Now().Truncate(time.Second).Second() == 59 {
				currentCandle = (*service.CandleInfraStruct)(candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute)))
				if currentCandle == nil {
					for i := 0; i < 5; i++ {
						currentCandle = (*service.CandleInfraStruct)(candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute)))
						if currentCandle != nil {
							break
						}
					}
				}
				fmt.Println("currentCandle:注文準備")
				fmt.Println(currentCandle)
				if currentCandle != nil {
					cross := currentCandle.Open / currentCandle.Close
					fmt.Println("cross")
					fmt.Println(cross)
					// 値幅が1000円以上の場合
					highToLow := currentCandle.High - currentCandle.Low
					fmt.Println("highToLow")
					fmt.Println(highToLow)
					params := map[string]string{
						"product_code":      "FX_BTC_JPY",
						"child_order_state": "ACTIVE",
					}
					orderRes, _ := bitflyerClient.ListOrder(params)
					// 十字線判定
					if len(orderRes) == 0 {
						fmt.Println("cross")
						fmt.Println(cross)
						//if (cross > 0.99994 && cross < 1.00006) || highToLow > 2000 {
						if (cross > 0.9999 && cross < 1.0001) || highToLow > 2000 {
							log.Println("currentCandle")
							log.Println(currentCandle)
							log.Println("十字線または設定値を超える値幅を検知しました。取引を2分休みます。")
							goto Pause
						}

						// 連続シグナル判定
						prev1Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute))
						prev2Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*1))
						prev3Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*2))
						prev4Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*3))
						prev5Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*4))

						if prev1Candle != nil && prev2Candle != nil && prev3Candle != nil && prev4Candle != nil && prev5Candle != nil {
							prev1UpperStatus := prev1Candle.Open < prev1Candle.Close
							prev2UpperStatus := prev2Candle.Open < prev2Candle.Close
							prev3UpperStatus := prev3Candle.Open < prev3Candle.Close
							prev4UpperStatus := prev4Candle.Open < prev4Candle.Close
							prev5UpperStatus := prev5Candle.Open < prev5Candle.Close
							fmt.Println("prev1UpperStatus")
							fmt.Println(prev1UpperStatus)
							fmt.Println("prev2UpperStatus")
							fmt.Println(prev2UpperStatus)
							fmt.Println("prev3UpperStatus")
							fmt.Println(prev3UpperStatus)
							fmt.Println("prev4UpperStatus")
							fmt.Println(prev4UpperStatus)
							fmt.Println("prev5UpperStatus")
							fmt.Println(prev5UpperStatus)
							if prev1UpperStatus == true && prev2UpperStatus == true && prev3UpperStatus == true && prev4UpperStatus == true && prev5UpperStatus == true {
								log.Println("同一のシグナルが連続で発生しているため取引を3分間中断します。")
								goto Pause
							} else if prev1UpperStatus == false && prev2UpperStatus == false && prev3UpperStatus == false && prev4UpperStatus == false && prev5UpperStatus == false {
								log.Println("同一のシグナルが連続で発生しているため取引を3分間中断します。")
								goto Pause
							}
						}
					}
				} else {
					log.Println("ローソク情報が取得できなかったため。2分間取引を中断します。")
					goto Pause
				}

				isUpper = service.IsUpperJudgment((*service.CandleInfraStruct)(currentCandle))
				log.Printf("isUpper（0：ロング, 1：ショート）：%s", strconv.Itoa(isUpper))
				closeOrderExecutionCheck = service.CloseOrderExecutionCheck()

				// 証拠金が設定範囲内か確認
				collateral, err := bitflyerClient.GetCollateral()
				i, _ := strconv.ParseFloat(os.Getenv("MIN_COLLATERAL"), 64)
				if err != nil {
					log.Fatalf("action=SystemTradeBase err=%s", err.Error())
				}
				if collateral.Collateral < i {
					fmt.Println(collateral)
					log.Fatal("証拠金が設定金額を下回っているため取引を中止します。")
				}
			}
		}
	}

Pause:
	for {
		for range time.Tick(1 * time.Second) {
			count++
			fmt.Println(count)
			if count == 175 {
				log.Println("Pause：システムトレードを再開します。")
				count = 0
				goto SystemTrade
			}
		}
	}

Mente:
	for {
		for range time.Tick(1 * time.Second) {
			menteCount++
			fmt.Println(menteCount)
			if menteCount == 2100 {
				log.Println("Mente：システムトレードを再開します。")
				menteCount = 0
				goto SystemTrade
			}
		}
	}

}
