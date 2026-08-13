package platform

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

// NewOracleDB opens a database/sql pool against an Oracle Database (ATP or
// ADW) via the godror driver and verifies connectivity with a ping. The DSN
// is a godror connect string ("user/password@tns_alias"); the wallet location
// is taken from the standard TNS_ADMIN environment variable. The caller owns
// the returned pool and must call Close() on shutdown.
//
// Note: godror requires Oracle Instant Client at runtime (build works
// without it) and CGO enabled at build time.
func NewOracleDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("godror", dsn)
	if err != nil {
		return nil, fmt.Errorf("open oracle connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping oracle: %w", err)
	}

	return db, nil
}
