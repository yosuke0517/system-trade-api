package service

import (
	"app/api/bitflyer"
	"app/infrastructure/databases/candle"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
	"unsafe"
)

type candleInfraStruct struct {
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
func SystemTradeService(productCode string, t time.Time) {

	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))

	// 証拠金が設定範囲内か確認
	collateral, err := bitflyerClient.GetCollateral()
	i, _ := strconv.ParseFloat(os.Getenv("MIN_COLLATERAL"), 64)
	if err != nil {
		log.Fatalf("action=SystemTradeBase err=%s", err.Error())
	}
	if collateral.Collateral < i {
		log.Fatal("証拠金が設定金額を下回っているため取引を中止します。")
	}

	// t := time.Now().Truncate(time.Second)
	fmt.Println("tttttt")
	fmt.Println(t.Truncate(time.Minute).Add(-time.Minute))
	// 0秒台で前回の分足ローソクを分析
	currentCandle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute))
	if currentCandle != nil {
		isUpper := isUpperJudgment((*candleInfraStruct)(currentCandle))
		fmt.Println(isUpper)
		if isUpper == true {
			// オープン注文
			fmt.Println("ロングー！！！！！！！")
			order := &bitflyer.Order{
				ProductCode:     "FX_BTC_JPY",
				ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
				Side:            "BUY",
				Size:            0.09, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
				MinuteToExpires: 1440,
				TimeInForce:     "GTC",
			}
			openRes, _ := bitflyerClient.SendOrder(order)
			// オープンが成功したら注文詳細を取得する（クローズ指値に使用する）
			time.Sleep(time.Second * 1)
			fmt.Println("openRes.ChildOrderAcceptanceID")
			fmt.Println(openRes.ChildOrderAcceptanceID)
			fmt.Println("unsafe.Sizeof(openRes)")
			fmt.Println(unsafe.Sizeof(openRes))
			if openRes.ChildOrderAcceptanceID == "" {
				log.Fatal("買付できない数量が指定されています。処理を終了します。")
			} else {
				params := map[string]string{
					"product_code":              "FX_BTC_JPY",
					"child_order_acceptance_id": openRes.ChildOrderAcceptanceID,
				}
				orderRes, _ := bitflyerClient.ListOrder(params)
				time.Sleep(time.Second * 2)
				fmt.Println("orderRes[0]")
				fmt.Println(orderRes[0])
				// クローズ注文
				// TODO 利益は要相談
				price := math.Floor(orderRes[0].AveragePrice * 1.00007)
				size := orderRes[0].Size
				// time.Sleep(time.Second * 1)
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
					fmt.Println(closeRes)
				}
			}
		}

		if isUpper == false {
			fmt.Println("ショート！！！！！！")
			// オープン注文
			order := &bitflyer.Order{
				ProductCode:     "FX_BTC_JPY",
				ChildOrderType:  "MARKET", // LIMIT(指値）or MARKET（成行）
				Side:            "SELL",
				Size:            0.11, // TODO フロントで計算する？？余計な計算入れたくないからフロントで計算したい
				MinuteToExpires: 1440,
				TimeInForce:     "GTC",
			}
			openRes, _ := bitflyerClient.SendOrder(order)
			// オープンが成功したら注文詳細を取得する（クローズ指値に使用する）
			time.Sleep(time.Second * 1)
			fmt.Println("openRes.ChildOrderAcceptanceID")
			fmt.Println(openRes.ChildOrderAcceptanceID)
			fmt.Println("unsafe.Sizeof(openRes)")
			fmt.Println(unsafe.Sizeof(openRes))
			if openRes.ChildOrderAcceptanceID == "" {
				log.Fatal("買付できない数量が指定されています。処理を終了します。")
			} else {
				params := map[string]string{
					"product_code":              "FX_BTC_JPY",
					"child_order_acceptance_id": openRes.ChildOrderAcceptanceID,
				}
				orderRes, _ := bitflyerClient.ListOrder(params)
				time.Sleep(time.Second * 2)
				fmt.Println("orderRes[0]")
				fmt.Println(orderRes[0])
				// クローズ注文
				// TODO 利益は要相談
				price := math.Floor(orderRes[0].AveragePrice * 0.99993)
				size := orderRes[0].Size
				// time.Sleep(time.Second * 1)
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

}

func isUpperJudgment(candle *candleInfraStruct) bool {
	// とりあえず陽線と陰線のみ
	fmt.Println("candle.Open")
	fmt.Println(candle.Open)
	fmt.Println("candle.Close")
	fmt.Println(candle.Close)
	if candle.Open < candle.Close {
		return true
	} else {
		return false
	}
}
