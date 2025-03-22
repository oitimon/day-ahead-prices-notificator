package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/loader"
	"github.com/shopspring/decimal"
	"log"
	"regexp"
	"time"
)

const measurementPrices = "prices"
const fieldNamePrice = "price"
const tagCurrency = "currency"

type Influx struct {
	ctx       context.Context
	cfg       *config.Influx
	client    influxdb2.Client
	writeAPI  api.WriteAPI
	queryAPI  api.QueryAPI
	deleteAPI api.DeleteAPI
	ldr       loader.Loader

	bucket            string
	measurementPrices string
	fieldNamePrice    string
}

func NewInflux(ctx context.Context, cfg *config.Influx, ldr loader.Loader) *Influx {
	inf := &Influx{
		ctx:    ctx,
		cfg:    cfg,
		client: influxdb2.NewClient(cfg.Url, cfg.Token),
		ldr:    ldr,
	}
	inf.bucket = inf.sanitizeInput(cfg.Bucket)
	inf.measurementPrices = inf.sanitizeInput(measurementPrices)
	inf.fieldNamePrice = inf.sanitizeInput(fieldNamePrice)

	inf.writeAPI = inf.client.WriteAPI(cfg.Orgname, cfg.Bucket)
	inf.queryAPI = inf.client.QueryAPI(cfg.Orgname)
	inf.deleteAPI = inf.client.DeleteAPI()

	errorsCh := inf.writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			log.Printf("Write error: %v", err)
		}
	}()

	// Close the client when the context is done.
	go func() {
		<-ctx.Done()
		log.Println("Closing InfluxDB client...")
		inf.client.Close()
	}()

	return inf
}

func (inf *Influx) Data() Data {
	return inf
}

func (inf *Influx) Bytes() Bytes {
	// @TODO: Implement this method
	return nil
}

func (inf *Influx) Get(startDate time.Time) (prices []decimal.Decimal, err error) {
	// Read from DB first.
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r._field == "%s")`,
		inf.bucket,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
		inf.measurementPrices,
		inf.fieldNamePrice,
	)
	log.Printf("Query to Influx: %s\n", query)
	// We create a sprintf-query as QueryWithParams is not working properly in the Go client.
	result, err := inf.queryAPI.Query(inf.ctx, query)
	if err != nil {
		err = errors.New("error querying data from Influx: " + err.Error())
		return
	}
	count := 0
	for result.Next() {
		count++
		value, ok := result.Record().Value().(float64)
		if !ok {
			log.Printf("Unexpected type for value: %T", result.Record().Value())
			log.Printf("Reset values and write them again...")
			count = 0
			break
		}
		prices = append(prices, decimal.NewFromFloat(value))
	}
	if result.Err() != nil {
		err = errors.New("error query parsing data from Influx: " + result.Err().Error())
		return
	}

	if count <= 0 {
		prices, err = inf.loadData(startDate)
	}
	return
}

func (inf *Influx) loadData(startDate time.Time) (prices []decimal.Decimal, err error) {
	// Clear existing data.
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
	if err = inf.deleteAPI.DeleteWithName(
		inf.ctx,
		inf.cfg.Orgname,
		inf.cfg.Bucket,
		startDate,
		endDate,
		fmt.Sprintf(`_measurement="%s"`, measurementPrices),
	); err != nil {
		err = errors.New("error deleting old data from Influx: " + err.Error())
		return
	}

	// Fetch data from the loader.
	prices, err = inf.ldr.Fetch(startDate)
	if err != nil {
		err = errors.New("error fetching prices: " + err.Error())
		return
	}

	// Write data to InfluxDB.
	for i := 0; i < len(prices); i++ {
		p := influxdb2.NewPoint(
			measurementPrices,
			map[string]string{"currency": tagCurrency},                         // tags
			map[string]interface{}{fieldNamePrice: prices[i].InexactFloat64()}, // fields
			startDate.Add(time.Hour*time.Duration(i)),                          // timestamp
		)

		inf.writeAPI.WritePoint(p)
	}
	inf.writeAPI.Flush()
	log.Printf("Writing %d prices to InfluxDB\n", len(prices))
	return
}

func (inf *Influx) sanitizeInput(input string) string {
	// Sanitize the input string to prevent injection attacks.
	// This is a simple example, you may want to use a more robust sanitization method.
	allowedChars := regexp.MustCompile(`[^a-zA-Z0-9 _\-.,@]`)
	return allowedChars.ReplaceAllString(input, "")
}
