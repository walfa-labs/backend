package platform

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx database/sql driver as "pgx"
)

// NewPostgresDB opens a database/sql pool against a PostgreSQL database via
// the pgx/v5 stdlib driver and verifies connectivity with a ping. The dsn is
// a standard PostgreSQL URL: "postgres://user:pass@host:port/dbname?sslmode=…".
// The caller owns the returned pool and must call Close() on shutdown.
func NewPostgresDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
