package repository

import (
	"app/domain/model"
	"time"
)

type CandleRepository interface {
	Insert() error
	Save() error
	GetCandle(productCode string, duration time.Duration, dateTime time.Time) (model.Candle, error)
	GetAllCandle(productCode string, duration time.Duration, limit int)
}
