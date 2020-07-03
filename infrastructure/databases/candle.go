package databases

import (
	api "app/api/bitflyer"
	"app/infrastructure"
	"fmt"
	"log"
	"time"
)

var location = time.FixedZone("Asia/Tokyo", 9*60*60)

type candleInfraStruct struct {
	ProductCode string
	Duration    time.Duration
	Time        time.Time
	Open        float64
	Close       float64
	High        float64
	Low         float64
	Volume      float64
}

func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open, close, high, low, volume float64) *candleInfraStruct {
	return &candleInfraStruct{
		ProductCode: productCode,
		Duration:    duration,
		Time:        timeDate,
		Open:        open,
		Close:       close,
		High:        high,
		Low:         low,
		Volume:      volume,
	}
}

// テーブルネームを取得する関数
func GetCandleTableName(productCode string, duration time.Duration) string {
	return fmt.Sprintf("%s_%s", productCode, duration)
}

// テーブルネームを取得するメソッド
func (c *candleInfraStruct) TableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

// キャンドル情報を追加する
func (c *candleInfraStruct) Insert() error {
	cmd := fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?)", c.TableName())
	ins, err := infrastructure.DB.Prepare(cmd)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("c.Time.In(location)")
	fmt.Println(c.Time.In(location))
	_, err = ins.Exec(c.Time.In(location), c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		log.Println(err)
	}
	return nil
}

// キャンドル情報を更新する
func (c *candleInfraStruct) Save() error {
	cmd := fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?", c.TableName())
	upd, err := infrastructure.DB.Prepare(cmd)
	if err != nil {
		log.Println(err)
	}
	upd.Exec(c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.In(location))
	return nil
}

// キャンドル情報を取得する
func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *candleInfraStruct {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf("SELECT time, open, close, high, low, volume FROM  %s WHERE time = ?", tableName)
	var candle candleInfraStruct
	err := infrastructure.DB.QueryRow(cmd, dateTime).Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil {
		return nil
	}
	return NewCandle(productCode, duration, candle.Time, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}

// キャンドル情報を保存する
func CreateCandleWithDuration(ticker api.Ticker, productCode string, duration time.Duration) bool {
	currentCandle := GetCandle(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	// 秒単位は毎回insertして欲しい
	if currentCandle == nil {
		candle := NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		fmt.Println("ticker.TruncateDateTime(duration)")
		fmt.Println(ticker.TruncateDateTime(duration))
		fmt.Println("duration")
		fmt.Println(duration)
		fmt.Println("candle")
		fmt.Println(&candle)
		candle.Insert()
		return true
	}
	// 分・時単位は秒単位ではupdateする
	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	currentCandle.Volume += ticker.Volume
	currentCandle.Close = price
	currentCandle.Save()
	return false
}
