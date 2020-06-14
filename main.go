package main

import (
	"app/bitflyer"
	"app/config"
	"app/utils"
	"fmt"
	"net/http"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("hoge")
	fmt.Fprintf(writer, "Hello World, %s!!", request.URL.Path[1:])
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	ticker, _ := apiClient.GetTicker("BTC_USD")
	fmt.Println(ticker.GetMidPrice())
	fmt.Println(apiClient.GetBalance())
	fmt.Println(apiClient.GetTicker("BTC_USD"))
	// v1/tickerのレスポンスはゾーン情報が含まれていないので変換できない
	//fmt.Println(ticker.DateTime())
	//fmt.Println(ticker.TruncateDateTime(time.Hour))
}

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
