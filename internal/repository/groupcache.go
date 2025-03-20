package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/loader"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// GroupCache is a global variable that holds the single groupcache instance
var gc *GroupCache = &GroupCache{}

type GroupCache struct {
	sync.Once

	data  groupCacheData
	bytes groupCacheBytes
}

type groupCacheData struct {
	cache  *groupcache.Group
	parser fastjson.Parser
}

type groupCacheBytes struct {
}

func NewGroupCache(cfg *config.GroupCache, ldr loader.Loader) *GroupCache {
	gc.Do(func() {
		peers := groupcache.NewHTTPPool(cfg.Me)
		if len(cfg.Peers) > 0 {
			peers.Set(cfg.Peers...)
		}

		gc.data = groupCacheData{
			cache: groupcache.NewGroup("data", 64<<20, groupcache.GetterFunc(
				func(ctx groupcache.Context, key string, dest groupcache.Sink) (err error) {
					log.Println("GroupCache, fetching from external source:", key)
					startDate, err := time.Parse(time.RFC3339, key)
					if err != nil {
						err = errors.New("error parsing date: " + err.Error())
						return
					}
					prices, err := ldr.Fetch(startDate)
					if err != nil {
						err = errors.New("error fetching prices: " + err.Error())
						return
					}
					value, err := json.Marshal(prices)
					if err != nil {
						err = errors.New("error marshalling prices: " + err.Error())
						return
					}
					if err = dest.SetBytes(value); err != nil {
						err = errors.New("error setting value bytes in groupcache: " + err.Error())
						return
					}
					return
				},
			)),
			parser: fastjson.Parser{},
		}

		gc.bytes = groupCacheBytes{}

		go func() {
			log.Println("Starting groupcache server on", cfg.Me)
			// Some default settings for the server.
			srv := &http.Server{
				Addr:              cfg.Listen,
				ReadHeaderTimeout: 15 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       30 * time.Second,
				Handler:           peers,
			}
			if err := srv.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()
	})
	return gc
}

func (gc *GroupCache) Data() Data {
	return &gc.data
}

func (gc *GroupCache) Bytes() Bytes {
	return &gc.bytes
}

func (gc *groupCacheData) Get(startDate time.Time) (prices []decimal.Decimal, err error) {
	// Implement the logic to get data from groupcache
	var data []byte
	if err = gc.cache.Get(nil, startDate.Format(time.RFC3339), groupcache.AllocatingByteSliceSink(&data)); err != nil {
		err = errors.New("error getting data from groupcache: " + err.Error())
		return
	}
	v, err := gc.parser.ParseBytes(data)
	if err != nil {
		err = fmt.Errorf("error parsing JSON '%s': %v", string(data), err)
		return
	}
	if v.Type() != fastjson.TypeArray {
		err = fmt.Errorf("unexpected JSON type '%s', expected array", v.Type())
		return
	}
	values := v.GetArray()
	for i := 0; i < len(values); i++ {
		price := values[i]
		if price.Type() != fastjson.TypeString {
			err = fmt.Errorf("unexpected JSON type '%s', expected string", price.Type())
			return
		}
		if priceFloat, err := strconv.ParseFloat(string(price.GetStringBytes()), 64); err != nil {
			err = fmt.Errorf("error parsing price '%s': %v", price.String(), err)
			return prices, err
		} else {
			prices = append(prices, decimal.NewFromFloat(priceFloat))
		}
	}
	return
}

func (gc *groupCacheBytes) Get(startDate time.Time) ([]byte, error) {
	// Implement the logic to get data from groupcache
	return nil, errors.New("not implemented")
}
