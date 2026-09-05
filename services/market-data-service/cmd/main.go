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
	svc "trading-platform/services/market-data-service/internal"
)

func main() {
	httpAddr := env("HTTP_ADDR", ":8080")
	databaseURL := env("DATABASE_URL", "postgres://dummy:dummy@localhost:5432/dummy?sslmode=disable")
	jwtSecret := env("JWT_SECRET", "local-dev-secret-change-me")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := dbx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := dbx.Migrate(ctx, pool, svc.MigrationsFS, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	tokens := authjwt.NewTokens(jwtSecret, 24*time.Hour)
	handler := svc.NewServer(svc.NewHandler(svc.NewRepository(pool)), tokens)
	httpServer := &http.Server{Addr: httpAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Printf("market-data-service listening on %s", httpAddr)
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

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
