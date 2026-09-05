package authjwt

import (
	"net/http"
	"strings"

	"trading-platform/pkg/web"

	"github.com/google/uuid"
)

func Authenticate(tokens *Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
				return
			}
			raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := tokens.Parse(raw)
			if err != nil {
				web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
				return
			}
			id, err := uuid.Parse(claims.Subject)
			if err != nil {
				web.WriteError(w, http.StatusUnauthorized, web.CodeUnauthorized, "Missing or invalid access token")
				return
			}
			ctx := web.WithPrincipal(r.Context(), web.Principal{ID: id, Email: claims.Email, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := web.PrincipalFrom(r.Context())
		if !ok || p.Role != "admin" {
			web.WriteError(w, http.StatusForbidden, web.CodeForbidden, "Admin role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
