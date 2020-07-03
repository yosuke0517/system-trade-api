package main

import (
	"app/application/controllers"
	"app/infrastructure"
	"app/utils"
	"fmt"
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

	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env")
	}

	//Middlewares hoge
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	utils.LoggingSettings(os.Getenv("LOG_FILE"))
	fmt.Println(infrastructure.DB)
	controllers.StreamIngestionData()
	// apiClient := api.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
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
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
	// オーダー一覧 TODO 固定じゃなくて動的にするfff
	//i := "JRF20200620-065843-055784"
	//params := map[string]string{
	//	"product_code":              "FX_BTC_JPY",
	//	"child_order_acceptance_id": i,
	//}
	//r, _ := apiClient.ListOrder(params) // TODO: s注文できなかったときはerrが返ってこなくて「""」で返ってくる
	//fmt.Println(r)

}
