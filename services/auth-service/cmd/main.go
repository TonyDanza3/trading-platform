package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trading-platform/pkg/authjwt"
	"trading-platform/pkg/dbx"
	svc "trading-platform/services/auth-service/internal"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := dbx.Connect(ctx, cfg.databaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := dbx.Migrate(ctx, pool, svc.MigrationsFS, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := svc.SeedAdmin(ctx, pool, cfg.adminEmail, cfg.adminPassword); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	tokens := authjwt.NewTokens(cfg.jwtSecret, cfg.jwtTTL)
	handler := svc.NewServer(svc.NewHandler(svc.NewRepository(pool), tokens))
	httpServer := &http.Server{Addr: cfg.httpAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Printf("auth-service listening on %s", cfg.httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

type config struct {
	httpAddr      string
	databaseURL   string
	jwtSecret     string
	jwtTTL        time.Duration
	adminEmail    string
	adminPassword string
}

func loadConfig() (config, error) {
	ttl, err := time.ParseDuration(env("JWT_TTL", "24h"))
	if err != nil {
		return config{}, err
	}
	cfg := config{
		httpAddr:      env("HTTP_ADDR", ":8080"),
		databaseURL:   env("DATABASE_URL", "postgres://dummy:dummy@localhost:5432/dummy?sslmode=disable"),
		jwtSecret:     env("JWT_SECRET", "local-dev-secret-change-me"),
		jwtTTL:        ttl,
		adminEmail:    env("ADMIN_EMAIL", "admin@example.com"),
		adminPassword: env("ADMIN_PASSWORD", "admin123"),
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
