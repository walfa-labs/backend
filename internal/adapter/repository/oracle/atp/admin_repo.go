package atp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/walfa-labs/backend/internal/domain"
)

// AdminRepo implements port.AdminRepo against Oracle ATP.
type AdminRepo struct {
	db *sql.DB
}

// NewAdminRepo constructs an AdminRepo bound to the given pool.
func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

// GetByUsername returns the admin user with the given username.
func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	var u domain.AdminUser
	err := r.db.QueryRowContext(ctx, `
		SELECT admin_user_id, username, password_hash, created_at
		FROM admin_users WHERE username = :1`, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// Upsert inserts or updates the admin user by username.
func (r *AdminRepo) Upsert(ctx context.Context, u *domain.AdminUser) error {
	_, err := r.db.ExecContext(ctx, `
		MERGE INTO admin_users u
		USING (SELECT :1 AS username FROM DUAL) src
		ON (u.username = src.username)
		WHEN MATCHED THEN
			UPDATE SET password_hash = :2
		WHEN NOT MATCHED THEN
			INSERT (admin_user_id, username, password_hash, created_at)
			VALUES (:3, :1, :2, :4)`,
		u.Username, u.PasswordHash, u.ID.String(), u.CreatedAt,
	)
	return err
}
