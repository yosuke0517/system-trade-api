package _bk

import "time"

// FxBtcJpy1m 売買のイベントを書き込む
type FxBtcJpy1m struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	Close  float64   `json:"close"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Volume float64   `json:"volume"`
}
