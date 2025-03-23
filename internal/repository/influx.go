package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/shopspring/decimal"
	"log"
	"time"
)

const measurementPrices = "prices"
const fieldNamePrice = "price"
const tagCurrency = "currency"
const currencyValue = "EUR"

type Influx struct {
	cfg               *config.Influx
	client            influxdb2.Client
	writeAPI          api.WriteAPI
	queryAPI          api.QueryAPI
	deleteAPI         api.DeleteAPI
	prev              Data
	deleteBeforeWrite bool
}

func NewInflux(ctx context.Context, cfg *config.Influx, prev Data) (inf *Influx, err error) {
	inf = &Influx{
		cfg:    cfg,
		client: influxdb2.NewClient(cfg.Url, cfg.Token),
		prev:   prev,
		// You can set deleteBeforeWrite to true if you want to delete the data before writing it.
		// This is useful if you want to overwrite the data in the database.
		// But take into account it will delete all fields with the same measurement and tags.
		// It's not recommended to use it (only for debugging purposes).
		deleteBeforeWrite: false,
	}
	inf.writeAPI = inf.client.WriteAPI(cfg.Orgname, cfg.Bucket)
	inf.queryAPI = inf.client.QueryAPI(cfg.Orgname)
	inf.deleteAPI = inf.client.DeleteAPI()

	errorsCh := inf.writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			log.Printf("Influxdb Writer error: %v", err)
		}
	}()

	// Close the client when the context is done.
	go func() {
		// Main server context, to close the client when the server is closed.
		<-ctx.Done()
		log.Println("Closing InfluxDB client...")
		inf.client.Close()
	}()

	return
}

func (*Influx) IsFinal() bool {
	return false
}

func (inf *Influx) Get(ctx context.Context, startDate time.Time) (prices []decimal.Decimal, err error) {
	// Read from DB first.
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r._field == "%s")`,
		inf.cfg.Bucket,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
		measurementPrices,
		fieldNamePrice,
	)
	log.Printf("Query to Influx: %s\n", query)
	// We create a sprintf-query as QueryWithParams is not working properly in the Go client.
	result, err := inf.queryAPI.Query(ctx, query)
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
		prices, err = inf.loadData(ctx, startDate)
	}
	return
}

func (inf *Influx) loadData(ctx context.Context, startDate time.Time) (prices []decimal.Decimal, err error) {
	if inf.deleteBeforeWrite {
		log.Printf("WARNING: Deleting data from InfluxDB for %s\n", startDate.Format("2006-01-02"))
		// Clear existing data (all fields).
		endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
		if err = inf.deleteAPI.DeleteWithName(
			ctx,
			inf.cfg.Orgname,
			inf.cfg.Bucket,
			startDate,
			endDate,
			fmt.Sprintf(`_measurement="%s"`, measurementPrices),
		); err != nil {
			err = errors.New("error deleting old data from Influx: " + err.Error())
			return
		}
	}

	// Fetch data from the loader.
	if inf.prev == nil {
		err = errors.New("previous repository not set for InfluxDB")
		return
	}
	prices, err = inf.prev.Get(ctx, startDate)
	if err != nil {
		err = errors.New("error fetching prices: " + err.Error())
		return
	}

	// Write data to InfluxDB.
	for i := 0; i < len(prices); i++ {
		p := influxdb2.NewPoint(
			measurementPrices,
			map[string]string{tagCurrency: currencyValue},                      // tags
			map[string]interface{}{fieldNamePrice: prices[i].InexactFloat64()}, // fields
			startDate.Add(time.Hour*time.Duration(i)),                          // timestamp
		)

		inf.writeAPI.WritePoint(p)
	}
	inf.writeAPI.Flush()
	log.Printf("Writing %d prices to InfluxDB\n", len(prices))
	return
}
