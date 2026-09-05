package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin(ctx context.Context, pool *pgxpool.Pool, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	tag, err := pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, display_name, role)
		VALUES ($1, $2, 'Admin', 'admin')
		ON CONFLICT (email) DO NOTHING
	`, email, string(hash))
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	if tag.RowsAffected() > 0 {
		log.Printf("seeded admin user %s", email)
	}
	return nil
}
