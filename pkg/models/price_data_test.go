package models

import (
	"github.com/shopspring/decimal"
	"reflect"
	"testing"
)

func TestPricesFloat64(t *testing.T) {
	// Initialize a PriceData instance with sample data
	priceData := &PriceData{
		Prices: []PriceDataEntry{
			{Price: decimal.NewFromFloat(100), ReadingDate: "2023-01-01"},
			{Price: decimal.NewFromFloat(200), ReadingDate: "2023-01-02"},
			{Price: decimal.NewFromFloat(300), ReadingDate: "2023-01-03"},
		},
	}

	// Call the PricesFloat64 method
	prices := priceData.PricesDecimal()

	// Define the expected result
	expected := []decimal.Decimal{decimal.NewFromFloat(100), decimal.NewFromFloat(200), decimal.NewFromFloat(300)}

	// Verify that the returned slice matches the expected values
	if !reflect.DeepEqual(prices, expected) {
		t.Errorf("PricesFloat64() = %v, want %v", prices, expected)
	}
}
