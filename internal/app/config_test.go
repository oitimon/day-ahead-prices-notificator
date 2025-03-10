package app

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"github.com/shopspring/decimal"
)

func TestConfigSelfCheck(t *testing.T) {
	cfg := generateTestConfig()

	// Perform self-check
	err := cfg.SelfCheck() // This should not panic or return an error
	assert.Nil(t, err)
}

func TestConfigExample(t *testing.T) {
	cfg := &ConfigApp{}
	err := godotenv.Load("../../.env.example")
	if err != nil {
		log.Fatalf("Error loading .env.example file: %v", err)
	}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Error processing environment variables: %v", err)
	}
	if err := cfg.SelfCheck(); err != nil {
		log.Fatalf("Error checking configuration: %v", err)
	}

}

func generateTestConfig() *ConfigApp {
	return &ConfigApp{
		Analytics: ConfigAnalytics{
			HighPrice: decimal.NewFromFloat(0.2),
			LowPrice:  decimal.NewFromFloat(0.1),
		},
		Loader: ConfigLoader{
			InclBtw: true,
			Driver:  "energyzero",
			API: ConfigAPI{
				Endpoint: "http://localhost:8080",
			},
		},
		Server: ConfigServer{
			Port: "8080",
		},
		Messenger: ConfigMessenger{
			Driver: "telegram",
			Telegram: ConfigTelegram{
				Token:  "test",
				ChatID: 123,
			},
		},
	}
}
