// Package postgres implements all repository and analytics port interfaces
// against a PostgreSQL database using database/sql with the pgx/v5 driver.
// SQL uses $1, $2, ... positional placeholders and native PostgreSQL types:
// UUID, BOOLEAN, TEXT, JSONB, TIMESTAMPTZ, and standard ANSI pagination
// (OFFSET $n LIMIT $n).
package postgres

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// inPlaceholders returns positional placeholders "$1, $2, $3" for count elements starting at start.
func inPlaceholders(count, start int) string {
	if count <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("$")
		b.WriteString(strconv.Itoa(start + i))
	}
	return b.String()
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows, allowing shared
// scan helpers to work with both single-row and multi-row query results.
type rowScanner interface {
	Scan(dest ...any) error
}

// nullStr unwraps a sql.NullString, returning "" for NULL values.
// In PostgreSQL (unlike Oracle), empty strings are stored as empty strings,
// so NULL truly means "no value was set".
func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// nullTime converts a nullable Oracle timestamp to *time.Time (nil when NULL).
func nullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// isUniqueViolation checks if an error indicates a unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates unique")
}
