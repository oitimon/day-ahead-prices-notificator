package app

import (
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
	cfg := generateTestConfig()
	cfg.Loader.Driver = loaderDriverEnergyZero
	cfg.Loader.API.Endpoint = server.URL

	data, err := FetchPrices(&cfg.Loader, time.Now())
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

	_, err := FetchPrices(generateFakeConfig(server.URL), time.Now())
	assert.Error(t, err)
}

func generateFakeConfig(serverUrl string) *ConfigLoader {
	return &ConfigLoader{
		InclBtw: true,
		Driver:  "test",
		API: ConfigAPI{
			Endpoint: serverUrl,
		},
	}
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
