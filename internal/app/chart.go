package app

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
	"github.com/shopspring/decimal"
	"math"
	"strings"
)

const (
	barChar = "█"
)

func ChartText(cfg *ConfigAnalytics, prices []decimal.Decimal) (message string, err error) {
	return drawLinesBarChartHtml(cfg, prices, 30, true)
}

func drawLinesBarChartHtml(cfg *ConfigAnalytics, prices []decimal.Decimal, width int, markDown bool) (message string, err error) {
	maxVal, minVal := prices[0], prices[0]
	for _, price := range prices {
		if maxVal.LessThan(price) {
			maxVal = price
		}
		if price.LessThan(minVal) {
			minVal = price
		}
	}

	scale := float64(0)
	if maxVal.InexactFloat64() != 0 {
		scale = math.Abs(maxVal.InexactFloat64()-minVal.InexactFloat64()) / float64(width)
	}
	if scale == 0 {
		scale = 1
	}

	for i, price := range prices {
		bar := strings.Repeat(barChar, int((price.InexactFloat64()-minVal.InexactFloat64()+scale)/scale))
		marker := ""
		markerFont := ""
		priceString := price.StringFixed(2)
		if markDown {
			if price.LessThanOrEqual(cfg.LowPrice) {
				marker = "_"
			} else if price.GreaterThan(cfg.HighPrice) {
				marker = "*"
			}
			priceString = strings.Replace(priceString, ".", "\\.", -1)
			markerFont = "`"
		}
		message += fmt.Sprintf("%s%02d:00%s %s %s%s%s\n", markerFont, i, markerFont, bar, marker, priceString, marker)
	}

	return
}

func drawASCIIBarChart(prices []decimal.Decimal, width, height int) (message string, err error) {
	fPrices := make([]float64, len(prices))
	for i, price := range prices {
		fPrices[i], _ = price.Float64()
	}

	message = asciigraph.Plot(fPrices, asciigraph.Width(width), asciigraph.Height(height))

	timeLabels := []string{"00", " ", "06", " ", "12", " ", "18", " ", "23"}
	axisX := "        " + strings.Join(timeLabels, "  ") + strings.Repeat(" ", width/2-5)
	message += "\n" + axisX

	return
}
