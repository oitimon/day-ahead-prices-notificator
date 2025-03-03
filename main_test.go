package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SourcePriceEntry struct {
	Price       float64 `json:"price"`
	ReadingDate string  `json:"readingDate"`
}

func TestFetchPrices(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"prices":[{"price":100.0}]}`))
			},
		),
	)
	defer server.Close()

	data, err := fetchPrices(server.URL)
	require.NoError(t, err)
	assert.NotEmpty(t, data.Prices)
	assert.Equal(t, 100.0, data.Prices[0].Price)
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

	_, err := fetchPrices(server.URL)
	assert.Error(t, err)
}

func TestCreateChart(t *testing.T) {
	priceData := &PriceData{
		Prices: []PriceDataEntry{{Price: 100.0}},
	}
	day := time.Now()

	r, err := createChart(priceData, day)
	require.NoError(t, err)
	assert.NotNil(t, r)
}

//func TestSendPriceMessage(t *testing.T) {
//	cfg := &Config{
//		TelegramToken:  "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I",
//		TelegramChatID: 76918703,
//		HighPrice:      decimal.NewFromFloat(150.0),
//		LowPrice:       decimal.NewFromFloat(50.0),
//	}
//	priceData := &PriceData{
//		Prices: []PriceDataEntry{{Price: 100.0}},
//	}
//	day := time.Now()
//	r := bytes.NewReader([]byte("PNG test"))
//
//	err := sendPriceMessage(cfg, priceData, day, r)
//	assert.NoError(t, err)
//}
//
//func TestSendTelegramMessage(t *testing.T) {
//	cfg := &Config{
//		TelegramToken:  "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I",
//		TelegramChatID: 76918703,
//	}
//	message := "Test message"
//
//	sendTelegramMessage(cfg, message)
//}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Welcome to DA price notificator!", rr.Body.String())
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "healthy", rr.Body.String())
}
