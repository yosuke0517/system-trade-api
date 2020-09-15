package controllers

import (
	"app/api/bitflyer"
	"app/application/response"
	"app/config"
	"app/domain/service"
	"app/infrastructure/databases/candle"
	"fmt"
	"github.com/markcheno/go-talib"
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
		for {
			if time.Now().Truncate(time.Second).Hour() == 4 {
				if time.Now().Truncate(time.Second).Minute() < 30 {
					log.Println("StreamIngestionData:4時〜4時30分までメンテナンスのため取引を中断します。")
					goto StreamIngestionDataMente
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
	StreamIngestionDataMente:
		for {
			for range time.Tick(1 * time.Second) {
				menteCount++
				fmt.Println("menteCount:StramIngestionData")
				fmt.Println(menteCount)
				if menteCount == 2000 {
					log.Println("StreamIngestionDataMente：ローソク足情報収集を再開します。")
					menteCount = 0
					break StreamIngestionDataMente
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
	var isUpper int
	var closeOrderExecutionCheck = false
	var count = 0
	var smallPauseCount = 0
	var menteCount = 0
	var trend int // 1:ロング, 2:ショート, 3:ローソク情報不足, 4:ロングsmall, 5:ショートsmall
	var newTrend int
	var isTrendChange = false
	var profitRate float64
	var targetBalance float64
	var currentBalance float64
SystemTrade:
	for {
		// 1秒タイマー
		for range time.Tick(1 * time.Second) {
			// TODO 4時台は取引しない（cronで制御？？）
			fmt.Println(time.Now().Truncate(time.Second))
			if time.Now().Truncate(time.Second).Hour() == 4 {
				if time.Now().Truncate(time.Second).Minute() < 30 {
					candle.Truncate()
					log.Println("4時〜4時40分までメンテナンスのため取引を中断します。")
					goto Mente
				}
			}
			// 0秒台で分析・システムトレードを走らせる
			if (time.Now().Truncate(time.Second).Minute()%10 == 9 || time.Now().Truncate(time.Second).Minute() == 9) && time.Now().Truncate(time.Second).Second() == 0 {
				currentCollateral, err := bitflyerClient.GetCollateral()
				if err != nil {
					fmt.Println("currentCollateral.Collateral")
					fmt.Println(currentCollateral)
					fmt.Println("targetBalance")
					fmt.Println(targetBalance)
					fmt.Println("現在残高が取れない")
				}
				if closeOrderExecutionCheck == true {
					go service.SystemTradeService(isUpper, profitRate)
					closeOrderExecutionCheck = false
				}
			}
			if (time.Now().Truncate(time.Second).Minute()%10 == 0 || time.Now().Truncate(time.Second).Minute() == 0) && time.Now().Truncate(time.Second).Second() == 5 {
				currentCollateral, err := bitflyerClient.GetCollateral()
				if err != nil {
					fmt.Println("currentCollateral.Collateral")
					fmt.Println(currentCollateral)
					fmt.Println("targetBalance")
					fmt.Println(targetBalance)
					fmt.Println("現在残高が取れない")
				}
				if closeOrderExecutionCheck == true {
					go service.SystemTradeService(isUpper, profitRate)
					closeOrderExecutionCheck = false
				}
			}
			//if time.Now().Truncate(time.Second).Second() % 30 == 0 && time.Now().Truncate(time.Second).Second() != 0 && time.Now().Truncate(time.Second).Second() != 60 {
			if (time.Now().Truncate(time.Second).Minute()%10 == 0 || time.Now().Truncate(time.Second).Minute() == 0) && time.Now().Truncate(time.Second).Second()%25 == 0 {
				closeOrderExecutionCheck = service.CloseOrderExecutionCheck()
				// isUpper, profitRate, isTrendChange = service.SmaAnalysis(isUpper, newTrend)
				currentCollateral, err := bitflyerClient.GetCollateral()
				if err != nil {
					fmt.Println("currentCollateral.Collateral")
					fmt.Println(currentCollateral)
					fmt.Println("targetBalance")
					fmt.Println(targetBalance)
					fmt.Println("現在残高が取れない")
				}
				if closeOrderExecutionCheck == true {
					go service.SystemTradeService(isUpper, profitRate)
					closeOrderExecutionCheck = false
				}
			}
			// ロスカット
			if time.Now().Truncate(time.Second).Second() == 55 {
				fmt.Println(isTrendChange)
				params := map[string]string{
					"product_code":      "FX_BTC_JPY",
					"child_order_state": "ACTIVE",
				}
				orderRes, _ := bitflyerClient.ListOrder(params)
				log.Println("orderRessssssss")
				log.Println(orderRes)
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
					fmt.Println("注文から120分以上経過したかどうか？")
					fmt.Println(orderTime.Add(time.Minute * 120).Before(time.Now()))
					execLossCut := service.LossCut(trend)
					log.Println("execLossCut")
					log.Println(execLossCut)
					// TODO 損切りの条件（仮）注文してから60分経過 or 注文時の価格と現在価格が2000円以上差がある時 ||中止中
					if orderTime.Add(time.Minute*30).Before(time.Now()) == true || math.Abs(limitPrice) > 4000 {
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
							log.Println(closeRes)
							if closeRes.ChildOrderAcceptanceID == "" {
								time.Sleep(time.Second * 1)
								for i := 0; i < 50; i++ {
									closeRes, _ := bitflyerClient.SendOrder(order)
									log.Println("closeRes")
									log.Println(closeRes.ChildOrderAcceptanceID)
									if closeRes.ChildOrderAcceptanceID != "" {
										break
									}
								}
							}
						}
						// 損切り後様子を見る
						goto Pause
					}
				}
			}

			// 注文準備
			if (time.Now().Truncate(time.Second).Minute()%10 == 9 || time.Now().Truncate(time.Second).Minute() == 9) && (time.Now().Truncate(time.Second).Second() == 28 || time.Now().Truncate(time.Second).Second() == 58) {
				params := map[string]string{
					"product_code":      "FX_BTC_JPY",
					"child_order_state": "ACTIVE",
				}

				orderRes, err := bitflyerClient.ListOrder(params)
				if err != nil {
					fmt.Println("注文が取得できませんでした。取り敢えずPause")
					goto SmallPause
				}
				// 注文が残っていたら準備しない
				if len(orderRes) == 0 {
					currentCandle := (*service.CandleInfraStruct)(candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute)))
					if currentCandle == nil {
						for i := 0; i < 10; i++ {
							currentCandle = (*service.CandleInfraStruct)(candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute)))
							fmt.Println("currentCandle")
							fmt.Println(currentCandle)
							time.Sleep(time.Second * 1)
							if currentCandle != nil {
								break
							}
						}
					}
					// 連続シグナル判定
					//prev1Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute))
					//prev2Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*1))
					//prev3Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*2))
					//prev4Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*3))
					//prev5Candle := candle.SelectOne(os.Getenv("PRODUCT_CODE"), time.Minute, time.Now().Truncate(time.Minute).Add(-time.Minute*4))

					//if prev1Candle != nil && prev2Candle != nil && prev3Candle != nil && prev4Candle != nil {
					//	prev1UpperStatus := prev1Candle.Open < prev1Candle.Close
					//	prev2UpperStatus := prev2Candle.Open < prev2Candle.Close
					//	prev3UpperStatus := prev3Candle.Open < prev3Candle.Close
					//	prev4UpperStatus := prev4Candle.Open < prev4Candle.Close
					//	prev5UpperStatus := prev5Candle.Open < prev5Candle.Close
					//	if prev1UpperStatus == true && prev2UpperStatus == true && prev3UpperStatus == true && prev4UpperStatus == true && prev5UpperStatus == true {
					//		log.Println("同一のシグナルが連続で発生しているため取引を3分間中断します。")
					//		goto SmallPause
					//	} else if prev1UpperStatus == false && prev2UpperStatus == false && prev3UpperStatus == false && prev4UpperStatus == false && prev5UpperStatus == false {
					//		log.Println("同一のシグナルが連続で発生しているため取引を3分間中断します。")
					//		goto SmallPause
					//	}
					//}

					//fmt.Println("currentCandle")
					//fmt.Println(currentCandle)
					//if currentCandle != nil {
					//	cross := currentCandle.Open / currentCandle.Close
					//	fmt.Println("cross")
					//	fmt.Println(cross)
					//	params := map[string]string{
					//		"product_code":      "FX_BTC_JPY",
					//		"child_order_state": "ACTIVE",
					//	}
					//	orderRes, _ := bitflyerClient.ListOrder(params)
					//	// 既存のオーダーがない場合、十字線判定
					//	if len(orderRes) == 0 {
					//		dfs100, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 100)
					//		fmt.Println("cross")
					//		fmt.Println(cross)
					//		//if (cross > 0.99994 && cross < 1.00006) || highToLow > 2000 {
					//		if cross > 0.9999 && cross < 1.0001 {
					//			log.Println("currentCandle")
					//			log.Println(currentCandle)
					//			log.Println("十字線または設定値を超える値幅を検知しました。取引を10分休みます。")
					//			goto SmallPause
					//		}
					//		fmt.Println("isUpper")
					//		fmt.Println(isUpper)
					//		trend, profitRate, isTrendChange = service.SmaAnalysis(isUpper, newTrend)
					//		isUpper = trend
					//		fmt.Println("isUpper")
					//		fmt.Println(isUpper)
					//		if len(dfs100.Closes()) == 100 {
					//			value100 := talib.Sma(dfs100.Closes(), 100)
					//			// 100分線と現在のキャンドルの乖離を求める
					//			disparation := value100[99] / currentCandle.Open
					//			neary := value100[99] / currentCandle.Open
					//			fmt.Println("neary")
					//			fmt.Println(neary)
					//			if neary > 0.999 && neary < 1.001 {
					//				fmt.Println("100分線と現在価格が近いため10分休みます")
					//				goto SmallPause
					//			}
					//			fmt.Println("disparation")
					//			fmt.Println(disparation)
					//			// ロング・ショートそれぞれ乖離が大きかったらPauseする
					//			if isUpper == 1 && disparation < 0.96 {
					//				log.Println("ロング：乖離幅が大きいためPauseします")
					//				goto SmallPause
					//			}
					//			if isUpper == 2 && disparation > 1.04 {
					//				log.Println("ショート：乖離幅が大きいためPauseします")
					//				goto SmallPause
					//			}
					//		}
					//		if isUpper == 3 {
					//			//goto Pause
					//		}
					//	}
					//}
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
				} else {
					log.Println("クローズオーダーありのため注文準備はしません。")
				}
			}
		}
	}

Pause:
	for {
		for range time.Tick(1 * time.Second) {
			count++
			fmt.Println(count)
			if count == 600 {
				log.Println("Pause：システムトレードを再開します。")
				count = 0
				goto SystemTrade
			}
		}
	}

SmallPause:
	for {
		for range time.Tick(1 * time.Second) {
			smallPauseCount++
			fmt.Println(smallPauseCount)
			if smallPauseCount == 120 {
				log.Println("smallPause：システムトレードを再開します。")
				smallPauseCount = 0
				goto SystemTrade
			}
		}
	}

Mente:
	for {
		for range time.Tick(1 * time.Second) {
			menteCount++
			fmt.Println(menteCount)
			if menteCount == 8600 {
				currentCollateral, err := bitflyerClient.GetCollateral()
				if err != nil {
					log.Println("現在の残高が取得できませんでした。")
				} else {
					currentBalance = currentCollateral.Collateral
					targetBalance = currentBalance * 1.04
					log.Println("今日のターゲット：")
					log.Println(targetBalance)
				}
				log.Println("Mente：システムトレードを再開します。")
				go StreamIngestionData()
				goto SystemTrade
			}
		}
	}

}
