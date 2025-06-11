package chart

import (
	"bytes"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/guptarohit/asciigraph"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/shopspring/decimal"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	barChar       = "█"
	barEmptyChar  = " "
	htmlPageTitle = "EPEX NL %s"
	FormatPage    = 1
	FormatIframe  = 2
)

type CommonChart struct {
	cfg *config.Ui
}

func NewChart(cfg *config.Ui) Chart {
	return &CommonChart{
		cfg: cfg,
	}
}

func (c *CommonChart) HtmlChart(prices []decimal.Decimal, day time.Time, format int) (html []byte, err error) {
	log.Printf("Generating HTML chart for: %s\n", day.Format("2006-01-02"))

	chartWidth := c.cfg.HtmlChart.Width
	chartHeight := c.cfg.HtmlChart.Height
	titleShift := c.cfg.HtmlChart.TitleShift
	fontSize := c.cfg.HtmlChart.Fontsize
	if format == FormatIframe {
		chartWidth = c.cfg.HtmlChart.IframeWidth
		chartHeight = c.cfg.HtmlChart.IframeHeight
		titleShift = c.cfg.HtmlChart.IframeTitleShift
		fontSize = c.cfg.HtmlChart.IframeFontsize
	}

	bar := charts.NewBar()
	xAxis := c.generateXAxis(prices)
	yAxis := c.generateYAxis(prices, fontSize)

	bar.SetXAxis(xAxis).
		AddSeries("", yAxis, charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(true)})).
		SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				Width:  chartWidth,
				Height: chartHeight,
			}),
			charts.WithTitleOpts(opts.Title{Title: fmt.Sprintf("EPEX NL %s", day.Format("2006-01-02")), Left: titleShift}),
			charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Rotate: 90, Formatter: opts.FuncOpts(`function (value) { return value.padStart(2, '0')+':00'; }`), FontSize: fontSize}}),
			charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Formatter: opts.FuncOpts(`function (value) { return value.toFixed(2); }`), FontSize: fontSize}}),
		).
		SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "inside"}))

	var buf bytes.Buffer
	if err = bar.Render(&buf); err != nil {
		return nil, fmt.Errorf("bar.Render(w): %w", err)
	}
	html = bytes.Replace(buf.Bytes(), []byte("Awesome go-echarts"), []byte(fmt.Sprintf(htmlPageTitle, c.cfg.Version)), 1)
	return
}

func (c *CommonChart) TextChart(prices []decimal.Decimal, day time.Time) (string, error) {
	log.Printf("Generating Text message for: %s\n", day.Format("2006-01-02"))
	return c.drawLinesBarChartMarkup(&c.cfg.Analytics, prices, c.cfg.TextChart.Width, true)
}

func (c *CommonChart) generateXAxis(prices []decimal.Decimal) []string {
	xAxis := make([]string, len(prices))
	for i := 0; i < len(prices); i++ {
		xAxis[i] = strconv.Itoa(i)
	}
	return xAxis
}

func (c *CommonChart) generateYAxis(prices []decimal.Decimal, fontSize int) []opts.BarData {
	yAxis := make([]opts.BarData, len(prices))
	for i, price := range prices {
		yAxis[i] = opts.BarData{
			Value: price.StringFixed(2),
			Label: &opts.Label{
				FontSize: float32(fontSize - 5),
			},
			ItemStyle: &opts.ItemStyle{
				Color: c.getColor(price, &c.cfg.Analytics),
			},
			Tooltip: &opts.Tooltip{
				Formatter: opts.FuncOpts(fmt.Sprintf(`function (params) {
				return '<div align="center">' + params.name.padStart(2, '0')+':00' + ' - ' + (Number(params.name)+1).toString().padStart(2, '0')+':00' + '<br/><b>' + %s + ' €' + '</b></span>';
			}`, price.StringFixed(4))),
			},
		}
	}
	return yAxis
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

func (c *CommonChart) drawLinesBarChartMarkup(cfg *config.Analytics, prices []decimal.Decimal, width int, markDown bool) (message string, err error) {
	maxVal, minVal := c.getMaxMinPrices(prices)
	scale := c.calculateScale(maxVal, minVal, width)

	for i, price := range prices {
		bar := ""
		if !price.IsNegative() && minVal.IsNegative() {
			bar = strings.Repeat(barEmptyChar, int((-minVal.InexactFloat64())/scale))
		}
		bar += strings.Repeat(barChar, int(math.Abs(price.InexactFloat64())/scale))
		marker, markerFont, priceString := c.getMarkerAndPriceString(price, cfg, markDown)
		message += fmt.Sprintf("%s%02d:00 %s %s%s%s%s\n", markerFont, i, bar, markerFont, marker, priceString, marker)
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

func (c *CommonChart) calculateScale(maxVal, minVal decimal.Decimal, width int) (scale float64) {
	if maxVal.IsZero() {
		maxVal = decimal.NewFromFloat(0.01)
	}
	if minVal.IsNegative() && !maxVal.IsNegative() {
		scale = math.Abs(maxVal.InexactFloat64()-minVal.InexactFloat64()) / float64(width)
	} else {
		scale = maxVal.InexactFloat64() / float64(width)
	}
	if scale == 0 {
		scale = 1
	}
	return
}

func (c *CommonChart) getMarkerAndPriceString(price decimal.Decimal, cfg *config.Analytics, markDown bool) (marker, markerFont, priceString string) {
	priceString = price.StringFixed(2)
	if markDown {
		if price.LessThanOrEqual(cfg.LowPrice) {
			marker = "_"
		} else if price.GreaterThan(cfg.HighPrice) {
			marker = "*"
		}
		//priceString = strings.Replace(priceString, ".", "\\.", -1)
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
