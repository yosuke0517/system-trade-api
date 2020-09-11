package main

import (
	"app/application/controllers"
	"app/application/server"
	"app/utils"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env")
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	e := echo.New()

	//Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	utils.LoggingSettings(os.Getenv("LOG_FILE"))

	/**
	リアルタイム controllerから
	*/
	go controllers.StreamIngestionData()
	go controllers.SystemTradeBase()
	//for range time.Tick(1 * time.Second) {
	//	dfs7, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 7)
	//	dfs14, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 14)
	//	dfs50, _ := service.GetAllCandle(os.Getenv("PRODUCT_CODE"), config.Config.Durations["1m"], 50)
	//	fmt.Println("df")
	//	fmt.Println(dfs7)
	//	// 各キャンドルのclose値を渡す
	//	value7 := talib.Sma(dfs7.Closes(), 7)
	//	value14 := talib.Sma(dfs14.Closes(), 14)
	//	value50 := talib.Sma(dfs50.Closes(), 50)
	//	fmt.Println(value7)
	//	fmt.Println(value14)
	//	fmt.Println(value50)
	//}
	/**
	APIClient
	*/
	// bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	// bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	/**
	資金確認
	*/
	// fmt.Println(bitflyerClient.GetBalance())
	// fmt.Println(bitflyerClient.GetCollateral())
	//
	/**
	リアルタイム apiから
	*/
	//tickerChannel := make(chan bitflyer.Ticker)
	//go bitflyerClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannel)
	//for ticker := range tickerChannel {
	//	fmt.Println(ticker)
	//	fmt.Println(ticker.GetMidPrice())
	//	fmt.Println(ticker.DateTime())
	//	fmt.Println(ticker.TruncateDateTime(time.Second))
	//	fmt.Println(ticker.TruncateDateTime(time.Minute))
	//	fmt.Println(ticker.TruncateDateTime(time.Hour))
	//}

	/**
	キャンドル情報取得
	*/
	// controllers.GetAllCandle()

	/**
	オーダー一覧 TODO 固定じゃなくて動的にする
	*/
	//i := "JRF20200710-160500-132315"
	//params := map[string]string{
	//	"product_code":              "FX_BTC_JPY",
	//	"child_order_acceptance_id": i,
	//}
	//r, _ := bitflyerClient.ListOrder(params) // TODO: s注文できなかったときはerrが返ってこなくて「""」で返ってくる
	//fmt.Println(r[0].AveragePrice)

	/**
	注文
	*/
	//order := &bitflyer.Order{
	//	ProductCode:     "FX_BTC_JPY",
	//	ChildOrderType:  "LIMIT",
	//	Side:            "BUY",
	//	Price:           800000,
	//	Size:            0.01,
	//	MinuteToExpires: 1440,
	//	TimeInForce:     "GTC",
	//}
	//res, _ := bitflyerClient.SendOrder(order)
	//fmt.Println(res)

	/**
	注文一覧
	*/
	//i := "JRF20200620-065843-055784"
	//params := map[string]string{
	//	"product_code":              "FX_BTC_JPY",
	//	"child_order_acceptance_id": i,
	//}
	//r, _ := bitflyerClient.ListOrder(params) // TODO: s注文できなかったときはerrが返ってこなくて「""」で返ってくる
	//fmt.Println(r[0])
	/**
	注文キャンセル
	*/
	//cancelOrder := &bitflyer.CancelOrder{
	//	ProductCode: "FX_BTC_JPY",
	//	ChildOrderAcceptanceID: "child_order_acceptance_id",
	//}
	//statusCode, _ := bitflyerClient.CancelOrder(cancelOrder)
	//fmt.Println(statusCode)
	server.Serve()
}
