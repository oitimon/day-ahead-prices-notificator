package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestConfigSelfCheck(t *testing.T) {
	cfg := GenerateTestConfig()

	// Perform self-check
	err := cfg.SelfCheck() // This should not panic or return an error
	assert.Nil(t, err)
}

func TestConfigExample(t *testing.T) {
	cfg := &App{}
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

func TestLocation(t *testing.T) {
	cfg := GenerateTestConfig()

	expected, _ := time.LoadLocation("Europe/Amsterdam")
	assert.Equal(t, expected, cfg.Location())
}

func TestTomorrowHourMin(t *testing.T) {
	cfg := GenerateTestConfig()

	assert.Equal(t, 15, cfg.TomorrowHourMin())
}
