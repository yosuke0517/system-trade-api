package main

import (
	"app/bitflyer"
	"app/config"
	"app/utils"
	"fmt"
	"net/http"
	"time"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	// not リアルタイム
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	ticker, _ := apiClient.GetTicker("BTC_USD")
	fmt.Println(ticker.GetMidPrice())
	fmt.Println(apiClient.GetBalance())
	fmt.Println(apiClient.GetTicker("BTC_USD"))

	// リアルタイム
	tickerChannel := make(chan bitflyer.Ticker)
	go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)
	for ticker := range tickerChannel {
		fmt.Println(ticker)
		fmt.Println(ticker.GetMidPrice())
		fmt.Println(ticker.DateTime())
		fmt.Println(ticker.TruncateDateTime(time.Second))
		fmt.Println(ticker.TruncateDateTime(time.Minute))
		fmt.Println(ticker.TruncateDateTime(time.Hour))
	}
	// v1/tickerのレスポンスはゾーン情報が含まれていないので変換できない
	//fmt.Println(ticker.DateTime())
	//fmt.Println(ticker.TruncateDateTime(time.Hour))
}

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
