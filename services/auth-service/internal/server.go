package internal

import (
	"net/http"

	"trading-platform/pkg/httpx"
	"trading-platform/pkg/web"

	"github.com/go-chi/chi/v5"
)

func NewServer(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.Recover)
	r.Use(httpx.RequestID)
	r.Use(httpx.CORS)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/api/v1/auth/register", h.Register)
	r.Post("/api/v1/auth/login", h.Login)
	return r
}
