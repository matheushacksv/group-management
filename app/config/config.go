package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	OpenWAURL         string
	OpenWAAdminAPIKey string
	OpenWASessionID   string
	WebhookToken      string
	APIKey            string
	SyncInterval      time.Duration
}

// Load get .env
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, err
	}

	port := os.Getenv("PORT")
	DBName := os.Getenv("POSTGRES_DB")
	DBUser := os.Getenv("POSTGRES_USER")
	DBPassword := os.Getenv("POSTGRES_PASSWORD")
	DBHost := os.Getenv("POSTGRES_HOST")
	DBPort := os.Getenv("POSTGRES_PORT")

	OpenWAURL := os.Getenv("OPENWA_BASEURL")
	OpenWAAdminKey := os.Getenv("OPENWA_ADMIN_APIKEY")
	OpenWASessionID := os.Getenv("OPENWA_SESSION_ID")
	WebhookToken := os.Getenv("WEBHOOK_TOKEN")
	APIKey := os.Getenv("API_KEY")

	if WebhookToken == "" {
		return Config{}, errors.New(
			"webhook token must be setted, generate a new token: openssl rand -hex 32",
		)
	}

	// empty key would match a request without the header and open the API
	if APIKey == "" {
		return Config{}, errors.New(
			"api key must be setted, generate a new key: openssl rand -hex 32",
		)
	}

	// auto sync interval, e.g. "6h" or "30m"; "0" disables it
	syncInterval := 6 * time.Hour
	if v := os.Getenv("SYNC_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SYNC_INTERVAL %q: %w", v, err)
		}
		syncInterval = d
	}

	databaseString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		DBUser, DBPassword, DBHost, DBPort, DBName,
	)

	if port == "" {
		port = "8080"
	}

	return Config{
		Port:              port,
		DatabaseURL:       databaseString,
		OpenWAURL:         OpenWAURL,
		OpenWAAdminAPIKey: OpenWAAdminKey,
		OpenWASessionID:   OpenWASessionID,
		WebhookToken:      WebhookToken,
		APIKey:            APIKey,
		SyncInterval:      syncInterval,
	}, nil
}
