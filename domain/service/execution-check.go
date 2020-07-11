package service

import (
	"app/api/bitflyer"
	"fmt"
	"os"
	"sync"
)

// クローズが約定しているかのチェック（建玉を保有していれば false, 保有していなければ true）
func CloseOrderExecutionCheck(closeOrderExecChannel chan bool, wg *sync.WaitGroup) {

	bitflyerClient := bitflyer.New(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	params := map[string]string{
		"product_code": "FX_BTC_JPY",
	}
	positionRes, _ := bitflyerClient.GetPositions(params)
	if len(positionRes) == 0 {
		fmt.Println("クローズオーダーなしのため取引可能")
		closeOrderExecChannel <- true
	} else {
		fmt.Println("クローズオーダーありのため取引不可")
		closeOrderExecChannel <- false
	}
	wg.Done()
}
