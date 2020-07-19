package service

import (
	"app/api/bitflyer"
	"fmt"
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
func SystemTradeService(isUpper int, currentCandle *CandleInfraStruct) {
	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	if isUpper == 0 {
		// オープン注文
		order := &bitflyer.Order{
			ProductCode:     "FX_BTC_JPY",
			ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
			Side:            "BUY",
			Size:            0.08, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
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
			price := math.Floor(orderRes[0].AveragePrice * 1.00004)
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

	if isUpper == 1 {
		// オープン注文
		order := &bitflyer.Order{
			ProductCode:     "FX_BTC_JPY",
			ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
			Side:            "SELL",
			Size:            0.1, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
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
			fmt.Println("orderRes[0]")
			fmt.Println(orderRes[0])
			// クローズ注文
			// TODO 利益は要相談
			price := math.Floor(orderRes[0].AveragePrice * 0.99996)
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
		return 1
	}

	lowerHige := prevCandle.Low / prevCandle.Close
	// 数値が小さいほどヒゲが長い
	if lowerHige < 0.99983 && prevCandle.Open < prevCandle.Close {
		log.Println("下ヒゲを検知しました。")
		log.Println(prevCandle)
		return 0
	}
	if prevCandle.Open < prevCandle.Close {
		return 0
	} else {
		return 1
	}
}
