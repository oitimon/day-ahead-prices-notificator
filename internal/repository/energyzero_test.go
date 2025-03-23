package repository

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchPrices(t *testing.T) {
	server := generateFakeServer()
	defer server.Close()
	cfg := &config.Energyzero{
		API: config.Api{
			Endpoint: server.URL,
		},
	}
	loader := NewEnergyzero(cfg)

	data, err := loader.Get(context.Background(), time.Now())
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Equal(t, float64(200), data[1].InexactFloat64())
}

func TestFetchPrices_Error(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		),
	)
	defer server.Close()
	cfg := &config.Energyzero{
		API: config.Api{
			Endpoint: server.URL,
		},
	}
	loader := NewEnergyzero(cfg)

	_, err := loader.Get(context.Background(), time.Now())
	assert.Error(t, err)
}

func TestIsFinal(t *testing.T) {
	e := &Energyzero{}
	assert.True(t, e.IsFinal())
}

func generateFakeServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"prices":[{"price":100.0},{"price":200.0}]}`))
			},
		),
	)
}
