package main

import (
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
	// go controllers.StreamIngestionData()

	/**
	APIClient
	*/
	// apiClient := api.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))

	/**
	リアルタイム apiから
	*/
	//tickerChannel := make(chan api.Ticker)
	//go apiClient.GetRealTimeTicker(os.Getenv("PRODUCT_CODE"), tickerChannel)
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
	//i := "JRF20200620-065843-055784"
	//params := map[string]string{
	//	"product_code":              "FX_BTC_JPY",
	//	"child_order_acceptance_id": i,
	//}
	//r, _ := apiClient.ListOrder(params) // TODO: s注文できなかったときはerrが返ってこなくて「""」で返ってくる
	//fmt.Println(r)

	/**
	注文
	*/
	//order := &api.Order{
	//	ProductCode:     "FX_BTC_JPY",
	//	ChildOrderType:  "LIMIT",
	//	Side:            "BUY",
	//	Price:           800000,
	//	Size:            0.01,
	//	MinuteToExpires: 1440,
	//	TimeInForce:     "GTC",
	//}
	//res, _ := apiClient.SendOrder(order)
	//fmt.Println(res)

	/**
	注文一覧
	*/
	//i := "JRF20200705-142019-326604"
	//params := map[string]string{
	//	"product_code":              "FX_BTC_JPY",
	//	"child_order_acceptance_id": i,
	//}
	// r, _ := apiClient.ListOrder(params) // TODO: s注文できなかったときはerrが返ってこなくて「""」で返ってくる

	/**
	注文キャンセル
	*/
	//cancelOrder := &api.CancelOrder{
	//	ProductCode: "FX_BTC_JPY",
	//	ChildOrderAcceptanceID: "child_order_acceptance_id",
	//}
	//statusCode, _ := apiClient.CancelOrder(cancelOrder)
	//fmt.Println(statusCode)
	server.Serve()
}
