package models

import (
	"github.com/shopspring/decimal"
	"sync"
)

type PriceDataEntry struct {
	Price       decimal.Decimal `json:"price"`
	ReadingDate string          `json:"readingDate"`
}

// PriceData struct to parse the response JSON
type PriceData struct {
	sync.Once

	Prices        []PriceDataEntry `json:"Prices"`
	pricesDecimal []decimal.Decimal
}

func (data *PriceData) PricesDecimal() []decimal.Decimal {
	data.Do(
		func() {
			data.pricesDecimal = make([]decimal.Decimal, len(data.Prices))
			for i, priceEntry := range data.Prices {
				data.pricesDecimal[i] = priceEntry.Price
			}
		},
	)
	return data.pricesDecimal
}
