package influx

import "time"

type PriceData struct {
	Timestamp    time.Time
	Price        float64
	PriceWithVat float64
}
