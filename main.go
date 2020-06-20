package main

import (
	"app/bitflyer"
	"app/config"
	"app/utils"
	"fmt"
	"net/http"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	// not リアルタイム
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	// ticker, _ := apiClient.GetTicker("BTC_JPY")
	//fmt.Print(ticker.GetMidPrice())
	//fmt.Print(apiClient.GetBalance())
	fmt.Print(apiClient.GetTicker("FX_BTC_JPY"))
	fmt.Println(apiClient.GetTradingCommission("FX_BTC_JPY"))

	// オーダー
	//order := &bitflyer.Order{
	//	ProductCode:     "FX_BTC_JPY", // TODO ここも動的に
	//	ChildOrderType:  "MARKET",  // TODO 動的にする {LIMIT 指値, MARKET 成り行き？
	//	Side:            "SELL",  // TODO 動的にする {BUY, SELL}
	//	// Price:           7000,  // TODO 動的にする LIMITのときの指定する値 (Option）
	//	Size:            0.01, // bitcoin数量
	//	MinuteToExpires: 1, // TODO 何コレ？？
	//	TimeInForce:     "GTC", // キャンセルするまで有効な注文
	//}
	//res, _ := apiClient.SendOrder(order)
	//fmt.Println(res.ChildOrderAcceptanceID)

	// リアルタイム
	//tickerChannel := make(chan bitflyer.Ticker)
	//go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)
	//for ticker := range tickerChannel {
	//	fmt.Println(ticker)
	//	fmt.Println(ticker.GetMidPrice())
	//	fmt.Println(ticker.DateTime())
	//	fmt.Println(ticker.TruncateDateTime(time.Second))
	//	fmt.Println(ticker.TruncateDateTime(time.Minute))
	//	fmt.Println(ticker.TruncateDateTime(time.Hour))
	//}
	// v1/tickerのレスポンスはゾーン情報が含まれていないので変換できない
	//fmt.Println(ticker.DateTime())
	//fmt.Println(ticker.TruncateDateTime(time.Hour))
}

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
	// オーダー一覧 TODO 固定じゃなくて動的にする
	i := "JRF20200620-065843-055784"
	params := map[string]string{
		"product_code":              "FX_BTC_JPY",
		"child_order_acceptance_id": i,
	}
	r, _ := apiClient.ListOrder(params) // TODO: 注文できなかったときはerrが返ってこなくて「""」で返ってくる
	fmt.Println(r)
}
