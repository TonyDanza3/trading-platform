package internal

import (
	"net/http"

	"trading-platform/pkg/authjwt"
	"trading-platform/pkg/httpx"
	"trading-platform/pkg/web"

	"github.com/go-chi/chi/v5"
)

func NewServer(h *Handler, tokens *authjwt.Tokens) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.AccessLog)
	r.Use(httpx.Recover)
	r.Use(httpx.RequestID)
	r.Use(httpx.CORS)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(authjwt.Authenticate(tokens))
		r.Get("/api/v1/users/me", h.GetMe)
		r.Patch("/api/v1/users/me", h.PatchMe)
		r.With(authjwt.RequireAdmin).Get("/api/v1/users", h.List)
		r.With(authjwt.RequireAdmin).Delete("/api/v1/users/{id}", h.Delete)
	})
	return r
}
