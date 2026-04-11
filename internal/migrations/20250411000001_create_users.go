package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upCreateUsers, downCreateUsers)
}

func upCreateUsers(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	return err
}

func downCreateUsers(ctx context.Context, db *bun.DB) error {
	_, err := db.NewDropTable().TableExpr("users").IfExists().Exec(ctx)
	return err
}
