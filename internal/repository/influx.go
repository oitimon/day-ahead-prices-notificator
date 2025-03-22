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
	"time"
)

const measurementPrices = "prices"
const tagCurrency = "currency"

type Influx struct {
	cfg       *config.Influx
	client    influxdb2.Client
	writeAPI  api.WriteAPI
	queryAPI  api.QueryAPI
	deleteAPI api.DeleteAPI
	ldr       loader.Loader
}

func NewInflux(ctx context.Context, cfg *config.Influx, ldr loader.Loader) *Influx {
	inf := &Influx{
		cfg:    cfg,
		client: influxdb2.NewClient(cfg.Url, cfg.Token),
		ldr:    ldr,
	}
	inf.writeAPI = inf.client.WriteAPI(cfg.Orgname, cfg.Bucket)
	inf.queryAPI = inf.client.QueryAPI(cfg.Orgname)
	inf.deleteAPI = inf.client.DeleteAPI()

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
			|> filter(fn: (r) => r._field == "price")
	`, inf.cfg.Bucket, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339), measurementPrices)
	log.Printf("Query to Influx: %s\n", query)
	result, err := inf.queryAPI.Query(context.Background(), query)
	if err != nil {
		err = errors.New("error querying data from Influx: " + err.Error())
		return
	}
	count := 0
	for result.Next() {
		count++
		prices = append(prices, decimal.NewFromFloat(result.Record().Value().(float64)))
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
		context.Background(),
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
			map[string]string{"currency": tagCurrency},                  // tags
			map[string]interface{}{"price": prices[i].InexactFloat64()}, // fields
			startDate.Add(time.Hour*time.Duration(i)),                   // timestamp
		)

		inf.writeAPI.WritePoint(p)
	}
	log.Printf("Writing %d prices to InfluxDB\n", len(prices))
	return
}
