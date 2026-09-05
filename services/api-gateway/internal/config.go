package internal

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr    string
	AuthURL     string
	UserURL     string
	MarketURL   string
	OpenAPIPath string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		AuthURL:     env("AUTH_URL", "http://localhost:8082"),
		UserURL:     env("USER_URL", "http://localhost:8083"),
		MarketURL:   env("MARKET_URL", "http://localhost:8084"),
		OpenAPIPath: env("OPENAPI_PATH", "api/openapi.yaml"),
	}
	for _, pair := range [][2]string{
		{"AUTH_URL", cfg.AuthURL},
		{"USER_URL", cfg.UserURL},
		{"MARKET_URL", cfg.MarketURL},
	} {
		if _, err := url.ParseRequestURI(pair[1]); err != nil {
			return Config{}, fmt.Errorf("%s: %w", pair[0], err)
		}
	}
	cfg.AuthURL = strings.TrimRight(cfg.AuthURL, "/")
	cfg.UserURL = strings.TrimRight(cfg.UserURL, "/")
	cfg.MarketURL = strings.TrimRight(cfg.MarketURL, "/")
	return cfg, nil
}

func (c Config) Routes() []Route {
	return []Route{
		{Prefix: "/api/v1/auth", Target: c.AuthURL},
		{Prefix: "/api/v1/users", Target: c.UserURL},
		{Prefix: "/api/v1/instruments", Target: c.MarketURL},
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
