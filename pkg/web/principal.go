package web

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const principalKey ctxKey = 1

type Principal struct {
	ID    uuid.UUID
	Email string
	Role  string
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}
