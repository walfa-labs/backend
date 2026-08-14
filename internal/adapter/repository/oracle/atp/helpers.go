// Package atp implements the driven repository adapters (port.*Repo
// interfaces) against Oracle Autonomous Transaction Processing via the godror
// driver and database/sql.
//
// Oracle-specific conventions used throughout the package:
//   - Placeholders are positional godror style (:1, :2, ...), not pgx $n.
//   - UUIDs are stored as VARCHAR2(36 CHAR); bind id.String(), scan into
//     uuid.UUID (its sql.Scanner accepts the string form).
//   - Booleans are NUMBER(1); use b2i/i2b — never bind a Go bool.
//   - Oracle stores ” as NULL, so every nullable text column scans through
//     sql.NullString (nullStr); there is no COALESCE(col, ”) trick here.
//   - CLOB columns scan into plain strings (godror fetches CLOBs as strings
//     by default) and are bound via clob() because plain string binds cap at
//     32767 bytes.
//   - experiences."current" is an Oracle reserved word and stays double-quoted
//     in every statement.
//   - No RETURNING clauses: inserts use the service-generated UUID and
//     re-SELECT created_at/updated_at by PK (DB DEFAULT CURRENT_TIMESTAMP).
package atp

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/godror/godror"
)

// inPlaceholders returns positional placeholders ":1, :2, :3" for count elements starting at start.
func inPlaceholders(count, start int) string {
	if count <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(":")
		b.WriteString(strconv.Itoa(start + i))
	}
	return b.String()
}

// rowScanner is satisfied by *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// b2i converts a Go bool to its Oracle NUMBER(1) representation.
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// i2b converts an Oracle NUMBER(1) value to a Go bool.
func i2b(i int) bool { return i != 0 }

// nullStr returns the string value of a sql.NullString ("" when NULL).
func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// nullTime converts a nullable Oracle timestamp to *time.Time (nil when NULL).
func nullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		t := nt.Time
		return &t
	}
	return nil
}

// clob binds a Go string to a CLOB column. Plain string binds max out at
// 32767 bytes; the Lob bind handles arbitrarily long markdown bodies.
func clob(s string) godror.Lob {
	return godror.Lob{Reader: strings.NewReader(s), IsClob: true}
}
