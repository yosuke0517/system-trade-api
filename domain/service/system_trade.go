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
	fmt.Println("分析する分足の日時")
	fmt.Println(t.Truncate(time.Minute).Add(-time.Minute))
	// 0秒台で前回の分足ローソクを分析
	currentCandle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute))
	fmt.Println("currentCandle.Open")
	if currentCandle == nil {
		for {
			currentCandle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute))
			if currentCandle != nil {
				break
			}
		}
	}
	fmt.Println(currentCandle.Open)
	//if currentCandle.High - currentCandle.Low > 2000 {
	//	var wg sync.WaitGroup
	//	wg.Add(1)
	//	go Pause(&wg, 600)
	//	wg.Wait()
	//}
	if currentCandle != nil {
		isUpper := IsUpperJudgment(productCode, t, (*CandleInfraStruct)(currentCandle))
		fmt.Println(isUpper)
		if isUpper == 0 {
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
			if openRes.ChildOrderAcceptanceID == "" {
				for {
					openRes, _ := bitflyerClient.SendOrder(order)
					if openRes.ChildOrderAcceptanceID != "" {
						break
					}
				}
			}
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
				if len(orderRes) == 0 {
					for {
						orderRes, _ = bitflyerClient.ListOrder(params)
						if len(orderRes) > 0 {
							break
						}
					}
				}
				fmt.Println("orderRes[0]")
				fmt.Println(orderRes[0])
				// クローズ注文
				// TODO 利益は要相談
				price := math.Floor(orderRes[0].AveragePrice * 1.00004)
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

		if isUpper == 1 {
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
			if openRes.ChildOrderAcceptanceID == "" {
				for {
					openRes, _ := bitflyerClient.SendOrder(order)
					if openRes.ChildOrderAcceptanceID != "" {
						break
					}
				}
			}
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
				if len(orderRes) == 0 {
					for {
						orderRes, _ = bitflyerClient.ListOrder(params)
						if len(orderRes) > 0 {
							break
						}
					}
				}
				fmt.Println("orderRes[0]")
				fmt.Println(orderRes[0])
				// クローズ注文
				// TODO 利益は要相談
				price := math.Floor(orderRes[0].AveragePrice * 0.99996)
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

		if isUpper == 2 {
			fmt.Println("何もしない")
		}
	}

}

func IsUpperJudgment(productCode string, t time.Time, prevCandle *CandleInfraStruct) int {
	prevcandle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute))
	// TODO 前回のローソク足が値幅2000円以上のとき5分間取引を中止する
	//fmt.Println("math.Abs(prevCandle.Close - prevCandle.Open)")
	//fmt.Println(math.Abs(prevCandle.Close - prevCandle.Open))
	//if math.Abs(prevCandle.Close-prevCandle.Open) > 2000 {
	//	return 2
	//}
	cross := 1.0 - (prevcandle.Open / prevcandle.Close)
	crossValue := math.Abs(cross)
	fmt.Println("crossValue")
	fmt.Println(crossValue)
	// 値幅が1000円以上の場合
	highToLow := prevcandle.High - prevcandle.Low
	fmt.Println("highToLow")
	fmt.Println(highToLow)
	if crossValue < 0.00005 || highToLow > 2000 {
		return 2
	}
	//prev1Candle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute*1))
	//prev2Candle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute*2))
	//prev3Candle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute*3))
	//prev4Candle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute*4))
	//prev5Candle := candle.SelectOne(productCode, time.Minute, t.Truncate(time.Minute).Add(-time.Minute*5))
	//
	//if prev1Candle != nil && prev2Candle != nil && prev3Candle != nil && prev4Candle != nil && prev5Candle != nil {
	//	prev1UpperStatus := prev1Candle.Open < prev1Candle.Close
	//	prev2UpperStatus := prev2Candle.Open < prev2Candle.Close
	//	prev3UpperStatus := prev3Candle.Open < prev3Candle.Close
	//	prev4UpperStatus := prev4Candle.Open < prev4Candle.Close
	//	prev5UpperStatus := prev5Candle.Open < prev5Candle.Close
	//	fmt.Println("prev1UpperStatus")
	//	fmt.Println(prev1UpperStatus)
	//	fmt.Println("prev2UpperStatus")
	//	fmt.Println(prev2UpperStatus)
	//	fmt.Println("prev3UpperStatus")
	//	fmt.Println(prev3UpperStatus)
	//	fmt.Println("prev4UpperStatus")
	//	fmt.Println(prev4UpperStatus)
	//	if prev1UpperStatus == true && prev2UpperStatus == true && prev3UpperStatus == true && prev4UpperStatus == true && prev5UpperStatus == true {
	//		return 2
	//	} else if prev1UpperStatus == false && prev2UpperStatus == false && prev3UpperStatus == false && prev4UpperStatus == false && prev5UpperStatus == false {
	//		return 2
	//	} else {
	//		if prevcandle.Open < prevcandle.Close {
	//			return 0
	//		} else {
	//			return 1
	//		}
	//	}
	//} else {
	if prevcandle.Open < prevcandle.Close {
		return 0
	} else {
		return 1
	}
	// }
}
