package chart

import (
	"bytes"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/guptarohit/asciigraph"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	barChar       = "█"
	htmlPageTitle = "EPEX NL %s"
)

type CommonChart struct {
	cfg *config.Analytics
}

func NewChart(cfg *config.Analytics) Chart {
	return &CommonChart{
		cfg: cfg,
	}
}

func (c *CommonChart) HtmlChart(prices []decimal.Decimal, day time.Time) (html []byte, err error) {
	log.Printf("Generating HTML chart for: %s\n", day.Format("2006-01-02"))

	bar := charts.NewBar()
	xAxis := c.generateXAxis(prices)
	yAxis := c.generateYAxis(prices)

	bar.SetXAxis(xAxis).
		AddSeries("", yAxis, charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(true)})).
		SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{Title: fmt.Sprintf("EPEX NL %s", day.Format("2006-01-02")), Left: "36%"}),
			charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Rotate: 90, Formatter: opts.FuncOpts(`function (value) { return value.padStart(2, '0')+':00'; }`)}}),
			charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Formatter: opts.FuncOpts(`function (value) { return value.toFixed(2); }`)}}),
			charts.WithTooltipOpts(opts.Tooltip{Formatter: opts.FuncOpts(`function (params) {
				return '<div align="center">' + params.name.padStart(2, '0')+':00' + ' - ' + (Number(params.name)+1).toString().padStart(2, '0')+':00' + '<br/><b>' + params.value + ' €' + '</b></span>';
			}`)}),
		).
		SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "inside"}))

	var buf bytes.Buffer
	if err = bar.Render(&buf); err != nil {
		return nil, fmt.Errorf("bar.Render(w): %w", err)
	}
	html = bytes.Replace(buf.Bytes(), []byte("Awesome go-echarts"), []byte(fmt.Sprintf(htmlPageTitle, c.cfg.Version)), -1)
	return
}

func (c *CommonChart) generateXAxis(prices []decimal.Decimal) []string {
	xAxis := make([]string, len(prices))
	for i := 0; i < len(prices); i++ {
		xAxis[i] = strconv.Itoa(i)
	}
	return xAxis
}

func (c *CommonChart) generateYAxis(prices []decimal.Decimal) []opts.BarData {
	yAxis := make([]opts.BarData, len(prices))
	for i, price := range prices {
		yAxis[i] = opts.BarData{
			Value: price,
			ItemStyle: &opts.ItemStyle{
				Color: c.getColor(price, c.cfg),
			},
		}
	}
	return yAxis
}

func (c *CommonChart) TextChart(prices []decimal.Decimal, day time.Time) (string, error) {
	log.Printf("Generating Text message for: %s\n", day.Format("2006-01-02"))
	return c.drawLinesBarChartHtml(c.cfg, prices, 30, true)
}

func (c *CommonChart) getColor(value decimal.Decimal, cfg *config.Analytics) string {
	if value.LessThanOrEqual(cfg.LowPrice) {
		return "green"
	} else if value.GreaterThanOrEqual(cfg.HighPrice) {
		return "red"
	} else {
		return ""
	}
}

func (c *CommonChart) drawLinesBarChartHtml(cfg *config.Analytics, prices []decimal.Decimal, width int, markDown bool) (message string, err error) {
	maxVal, minVal := c.getMaxMinPrices(prices)
	scale := c.calculateScale(maxVal, minVal, width)

	for i, price := range prices {
		bar := strings.Repeat(barChar, int((price.InexactFloat64()-minVal.InexactFloat64()+scale)/scale))
		marker, markerFont, priceString := c.getMarkerAndPriceString(price, cfg, markDown)
		message += fmt.Sprintf("%s%02d:00%s %s %s%s%s\n", markerFont, i, markerFont, bar, marker, priceString, marker)
	}

	return
}

func (c *CommonChart) getMaxMinPrices(prices []decimal.Decimal) (maxVal, minVal decimal.Decimal) {
	maxVal, minVal = prices[0], prices[0]
	for _, price := range prices {
		if maxVal.LessThan(price) {
			maxVal = price
		}
		if price.LessThan(minVal) {
			minVal = price
		}
	}
	return
}

func (c *CommonChart) calculateScale(maxVal, minVal decimal.Decimal, width int) float64 {
	scale := float64(0)
	if maxVal.InexactFloat64() != 0 {
		scale = math.Abs(maxVal.InexactFloat64()-minVal.InexactFloat64()) / float64(width)
	}
	if scale == 0 {
		scale = 1
	}
	return scale
}

func (c *CommonChart) getMarkerAndPriceString(price decimal.Decimal, cfg *config.Analytics, markDown bool) (marker, markerFont, priceString string) {
	priceString = price.StringFixed(2)
	if markDown {
		if price.LessThanOrEqual(cfg.LowPrice) {
			marker = "_"
		} else if price.GreaterThan(cfg.HighPrice) {
			marker = "*"
		}
		priceString = strings.Replace(priceString, ".", "\\.", -1)
		markerFont = "`"
	}
	return
}

func (c *CommonChart) drawASCIIBarChart(prices []decimal.Decimal, width, height int) (message string, err error) {
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
