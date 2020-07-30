package service

import (
	"app/api/bitflyer"
	"app/config"
	"fmt"
	"github.com/markcheno/go-talib"
	"log"
	"math"
	"os"
	"time"
)

type CandleInfraStruct struct {
	ProductCode string
	Duration    time.Duration
	Time        time.Time
	Open        float64
	Close       float64
	High        float64
	Low         float64
	Volume      float64
}

type Trade struct {
	isTrading bool
}

//○ 証拠金が10万円以下の場合はアラートを出してシステム終了
//○ 資金のチェック
//トレード中channelを受け取った場合は何もしない
//トレード可能channelを受け取った場合
//TODO 2 スプレッドのチェック
//0.01%より上の場合は1.へ戻る
//4 値動きチェック
//1分ローソク足のisUpperがtrueかどうか
//5 trueの場合
//5-1 ロング注文を出す（SendOrder）←作成済み
//5-2 注文テーブルに書き込み
//5-3 約定したらisTradingフラグを立て、利益確定価格を約定価格の0.00007%に設定し注文。それぞれのchild_order_acceptance_idをLossCutへ渡しロスカットジョブを走らせる。
//5-4. クローズ注文のchild_order_acceptance_idをCloseOrderExecutionCheckへ渡して約定したかどうかを監視する
//5 falseの場合ショートで5-1~5-4を実施。
func SystemTradeService(isUpper int, profitRate float64) {
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	log.Println("isUpper:SystemTradeService")
	log.Println(isUpper)
	if isUpper == 1 {
		// オープン注文
		order := &bitflyer.Order{
			ProductCode:     "FX_BTC_JPY",
			ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
			Side:            "BUY",
			Size:            0.07, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
			MinuteToExpires: 1440,
			TimeInForce:     "GTC",
		}
		openRes, _ := bitflyerClient.SendOrder(order)
		// オープンが成功したら注文詳細を取得する（クローズ指値に使用する）
		if openRes == nil {
			for i := 0; i < 10; i++ {
				openRes, _ := bitflyerClient.SendOrder(order)
				if openRes != nil {
					break
				}
			}
		}
		if openRes == nil {
			log.Fatal("オープンの注文が約定できませんでした。アプリケーションを終了します。")
		}
		if openRes.ChildOrderAcceptanceID == "" {
			log.Fatal("オープンの注文が約定できませんでした。アプリケーションを終了します。")
		} else {
			params := map[string]string{
				"product_code":              "FX_BTC_JPY",
				"child_order_acceptance_id": openRes.ChildOrderAcceptanceID,
			}
			orderRes, _ := bitflyerClient.ListOrder(params)
			if len(orderRes) == 0 {
				for i := 0; i < 50; i++ {
					orderRes, _ = bitflyerClient.ListOrder(params)
					if len(orderRes) > 0 {
						break
					}
					if i == 50 {
						if len(orderRes) == 0 {
							for i := 0; i < 50; i++ {
								time.Sleep(time.Second * 1)
								orderRes, _ = bitflyerClient.ListOrder(params)
								if len(orderRes) > 0 {
									break
								}
							}
						}
					}
				}
			}
			if len(orderRes) == 0 {
				log.Fatal("オープン注文が約定しませんでした。アプリケーションを終了します。")
			}
			// クローズ注文
			// TODO 利益は要相談
			price := math.Floor(orderRes[0].AveragePrice * profitRate)
			size := orderRes[0].Size
			if orderRes != nil {
				order := &bitflyer.Order{
					ProductCode:     "FX_BTC_JPY",
					ChildOrderType:  "LIMIT",
					Side:            "SELL",
					Price:           price,
					Size:            size,
					MinuteToExpires: 1440,
					TimeInForce:     "GTC",
				}
				closeRes, _ := bitflyerClient.SendOrder(order)
				log.Println("closeRes")
				log.Println(closeRes)
			}
		}
	}

	if isUpper == 2 {
		// オープン注文
		order := &bitflyer.Order{
			ProductCode:     "FX_BTC_JPY",
			ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
			Side:            "SELL",
			Size:            0.07, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
			MinuteToExpires: 1440,
			TimeInForce:     "GTC",
		}
		openRes, _ := bitflyerClient.SendOrder(order)
		// オープンが成功したら注文詳細を取得する（クローズ指値に使用する）
		if openRes == nil {
			for i := 0; i < 10; i++ {
				openRes, _ := bitflyerClient.SendOrder(order)
				if openRes != nil {
					break
				}
			}
		}
		if openRes == nil {
			log.Fatal("買付できない数量が指定されています。処理を終了します。")
		} else {
			params := map[string]string{
				"product_code":              "FX_BTC_JPY",
				"child_order_acceptance_id": openRes.ChildOrderAcceptanceID,
			}
			orderRes, _ := bitflyerClient.ListOrder(params)

			if len(orderRes) == 0 {
				for i := 0; i < 50; i++ {
					orderRes, _ = bitflyerClient.ListOrder(params)
					if len(orderRes) > 0 {
						break
					}
				}
			}
			if len(orderRes) == 0 {
				for i := 0; i < 50; i++ {
					orderRes, _ = bitflyerClient.ListOrder(params)
					if len(orderRes) > 0 {
						break
					}
				}
			}
			if len(orderRes) == 0 {
				for i := 0; i < 50; i++ {
					orderRes, _ = bitflyerClient.ListOrder(params)
					if len(orderRes) > 0 {
						break
					}
				}
				log.Fatal("オープン注文が約定しませんでした。アプリケーションを終了します。")
			}
			fmt.Println("orderRes[0]")
			fmt.Println(orderRes[0])
			// クローズ注文
			// TODO 利益は要相談
			price := math.Floor(orderRes[0].AveragePrice * profitRate)
			size := orderRes[0].Size
			if orderRes != nil {
				order := &bitflyer.Order{
					ProductCode:     "FX_BTC_JPY",
					ChildOrderType:  "LIMIT",
					Side:            "BUY",
					Price:           price,
					Size:            size,
					MinuteToExpires: 1440,
					TimeInForce:     "GTC",
				}
				closeRes, _ := bitflyerClient.SendOrder(order)
				fmt.Println(closeRes)
			}
		}
	}
}

func IsUpperJudgment(prevCandle *CandleInfraStruct) int {
	upperHige := prevCandle.High / prevCandle.Close
	// 数値が大きいほどヒゲが大きい
	if upperHige > 1.00017 && prevCandle.Open > prevCandle.Close {
		log.Println("上ヒゲを検知しました。")
		log.Println(prevCandle)
		return 2
	}

	lowerHige := prevCandle.Low / prevCandle.Close
	// 数値が小さいほどヒゲが長い
	if lowerHige < 0.99983 && prevCandle.Open < prevCandle.Close {
		log.Println("下ヒゲを検知しました。")
		log.Println(prevCandle)
		return 1
	}
	if prevCandle.Open < prevCandle.Close {
		return 1
	} else {
		return 2
	}
}

// 与えられたperiodに対するSMA値を返す // trend 0:ロング、1:ショート
// return int:ロング or ショート(1:ロング、2:ショート）, float64:クローズオーダーの率（トレンドによって変える）, bool:前回とトレンドが変わったかどうか
// 前回のトレンドを受け取りトレンドの変化を判定
func SmaAnalysis(trend, newTrend int) (int, float64, bool) {
	var profitRate = 0.0006
	dfs10, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 11)
	dfs21, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 21)
	// dfs100, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 70)
	// dfs45, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 45)
	if len(dfs21.Closes()) == 21 {
		// 各キャンドルのclose値を渡す
		value10 := talib.Sma(dfs10.Closes(), 11)
		value21 := talib.Sma(dfs21.Closes(), 21)
		// value100 := talib.Sma(dfs100.Closes(), 70)
		// value45 := talib.Sma(dfs45.Closes(), 45)
		// fmt.Println("value100")
		// fmt.Println(value100[69])
		// fmt.Println("value45")
		// fmt.Println(value45[44])
		fmt.Println("value21")
		fmt.Println(int(value21[20]))
		fmt.Println("value10")
		fmt.Println(int(value10[10]))
		// ロングトレンド
		//if value10[9] > value100[69] && value21[20] > value100[69] {
		//	log.Println("ロングトレンド")
		//	log.Println("value100")
		//	log.Println(value100)
		//	log.Println("value21")
		//	log.Println(value21)
		//	log.Println("value10")
		//	log.Println(value10)
		//	newTrend = 1
		//}
		if value10[10] > value21[20] {
			log.Println("ロングトレンド")
			// log.Println("value100")
			// log.Println(value100)
			fmt.Println("value21")
			fmt.Println(int(value21[20]))
			fmt.Println("value10")
			fmt.Println(int(value10[10]))
			newTrend = 1
		}
		// 7分平均のみロングへ移行した状態
		//if value10[9] > value100[69] && value21[20] < value100[69] {
		//	log.Println("ロングトレンドsmall")
		//	newTrend = 4
		//}
		//
		//// 7分平均のみショートへ移行した状態
		//if value10[9] < value100[69] && value21[20] > value100[69] {
		//	log.Println("ショートトレンドsmall")
		//	newTrend = 5
		//}

		// ショートトレンド
		//if value10[9] < value100[69] && value21[20] < value100[69] {
		//	log.Println("ショートトレンド")
		//	log.Println("value100")
		//	log.Println(value100)
		//	log.Println("value21")
		//	log.Println(value21)
		//	log.Println("value10")
		//	log.Println(value10)
		//	newTrend = 2
		//}
		if value10[10] < value21[20] {
			log.Println("ショートトレンド")
			// log.Println("value100")
			// log.Println(value100)
			fmt.Println("value21")
			fmt.Println(int(value21[20]))
			fmt.Println("value10")
			fmt.Println(int(value10[10]))
			newTrend = 2
		}
		fmt.Println("trend：")
		fmt.Println(trend)
		fmt.Println("newTrend：")
		fmt.Println(newTrend)
		// トレンドを検知したらisTrendChangeをtrueにする
		if trend != 0 && trend != 3 && trend != newTrend && newTrend != 0 {
			log.Println("Lトレンドの変更を検知しました。")
			log.Println(newTrend)
			if newTrend == 1 {
				return newTrend, 1.0 + profitRate, true
			}
			if newTrend == 2 {
				return newTrend, 1.0 - profitRate, true
			}
		}
		// 2回目以降でトレンドの変更がなかった場合はisTrendChangeはfalse
		if trend != 0 && trend == newTrend {
			return newTrend, 0, false
		}
	} else {
		log.Println("キャンドル数がトレード必要数に達していません。3分間取引を中断して必要なキャンドル情報を収集します。")
		return 3, 0, false
	}
	// 初回はisTrendChangeはfalseとする
	if trend == 0 {
		return newTrend, 0, false
	}
	return newTrend, 0, false
}

// トレンドのみ返す
func SimpleSmaAnalysis() int {
	var trend = 0
	dfs10, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 11)
	dfs21, _ := GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 21)
	if len(dfs21.Closes()) == 21 {
		// 各キャンドルのclose値を渡す
		value10 := talib.Sma(dfs10.Closes(), 11)
		value21 := talib.Sma(dfs21.Closes(), 21)
		fmt.Println("value21")
		fmt.Println(int(value21[20]))
		fmt.Println("value10")
		fmt.Println(int(value10[10]))
		// ロングトレンド
		if value10[10] > value21[20] {
			log.Println("ロングトレンド")
			// log.Println("value100")
			// log.Println(value100)
			fmt.Println("value21")
			fmt.Println(int(value21[20]))
			fmt.Println("value10")
			fmt.Println(int(value10[10]))
			trend = 1
			return trend
		}
		// ショートトレンド
		if value10[10] < value21[20] {
			log.Println("ショートトレンド")
			// log.Println("value100")
			// log.Println(value100)
			fmt.Println("value21")
			fmt.Println(int(value21[20]))
			fmt.Println("value10")
			fmt.Println(int(value10[10]))
			trend = 2
			return trend
		} else {
			log.Println("キャンドル数がトレード必要数に達していません。3分間取引を中断して必要なキャンドル情報を収集します。SimpleSmaAnalysis")
			return 3
		}
	}
	return trend
}
