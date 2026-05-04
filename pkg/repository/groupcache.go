package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
)

// GroupCache is a global variable that holds the single groupcache instance
var gc *GroupCache = &GroupCache{}

type GroupCache struct {
	sync.Once

	data  groupCacheData
	bytes groupCacheBytes
}

type groupCacheData struct {
	parent *GroupCache
	cache  *groupcache.Group
	prev   Data
}

type groupCacheBytes struct {
}

func NewGroupCache(cfg *config.GroupCache, prev Data) *GroupCache {
	gc.Do(func() {
		peers := groupcache.NewHTTPPool(cfg.Me)
		if len(cfg.Peers) > 0 {
			peers.Set(cfg.Peers...)
		}

		gc.data = groupCacheData{
			parent: gc,
			prev:   prev,
			cache: groupcache.NewGroup("data", 64<<20, groupcache.GetterFunc(
				func(ctx groupcache.Context, key string, dest groupcache.Sink) (err error) {
					log.Println("Groupcache, fetching from external source:", key)
					startDate, opts, err := gc.deserializeKey(key)
					if err != nil {
						err = errors.New("error parsing date: " + err.Error())
						return
					}
					if gc.data.prev == nil {
						err = errors.New("previous repository not set for Groupcache")
						return
					}
					prices, err := gc.data.prev.Get(ctx, startDate, opts...)
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

func (gc *GroupCache) serializeKey(startDate time.Time, options *Options) string {
	key := startDate.Format(time.RFC3339)
	if options.WithVat {
		key += "_wVat"
	}
	return key
}

func (gc *GroupCache) deserializeKey(key string) (startDate time.Time, opts []Option, err error) {
	parts := strings.Split(key, "_")
	if len(parts) < 1 {
		err = fmt.Errorf("invalid key format: '%s'", key)
		return
	}
	if startDate, err = time.Parse(time.RFC3339, parts[0]); err != nil {
		err = fmt.Errorf("error parsing date from key '%s': %v", key, err)
		return
	}
	if len(parts) > 1 && parts[1] == "wVat" {
		opts = append(opts, WithVat(true))
	}
	return
}

func (*groupCacheData) IsFinal() bool {
	return false
}

func (gc *groupCacheData) Get(ctx context.Context, startDate time.Time, opts ...Option) (prices []decimal.Decimal, err error) {
	options := NewOptions(opts...)
	var data []byte
	if err = gc.cache.Get(ctx, gc.parent.serializeKey(startDate, options), groupcache.AllocatingByteSliceSink(&data)); err != nil {
		err = errors.New("error getting data from groupcache: " + err.Error())
		return
	}
	v, err := fastjson.ParseBytes(data)
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

func (gc *groupCacheBytes) Get(ctx context.Context, startDate time.Time, opts ...Option) ([]byte, error) {
	// Implement the logic to get data from groupcache
	return nil, errors.New("not implemented")
}

func (*groupCacheBytes) IsFinal() bool {
	return false
}
