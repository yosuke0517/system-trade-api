package repository

import (
	"app/domain/model"
	"time"
)

type CandleRepository interface {
	Insert(time.Time, float64, float64, float64, float64, float64) error
	Save(time.Time, float64, float64, float64, float64, float64) error
	FindOne(time.Time) (model.Candle, error)
}
