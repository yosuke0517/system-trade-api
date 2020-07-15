package service

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var count int
var s string

func Pause(wg *sync.WaitGroup, limit int) {
	fmt.Println("設定値を超える分足の幅を検出しました。" + string(limit) + "秒間トレードを停止します。")
	for range time.Tick(1 * time.Second) {
		count++
		s = strconv.Itoa(limit - count)
		fmt.Println("取引再開まであと" + s + "秒")
		if count == limit {
			wg.Done()
		}
	}
}
