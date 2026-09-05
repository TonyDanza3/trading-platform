package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gw "trading-platform/services/api-gateway/internal"
)

func main() {
	cfg, err := gw.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	srv, err := gw.New(cfg)
	if err != nil {
		log.Fatalf("gateway: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on %s auth=%s user=%s market=%s", cfg.HTTPAddr, cfg.AuthURL, cfg.UserURL, cfg.MarketURL)
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
