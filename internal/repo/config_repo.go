package repo

import (
	"context"
	"database/sql"

	"mangav5/internal/models"
)

type ConfigRepo struct {
	DB *sql.DB
}

func NewConfigRepo(db *sql.DB) *ConfigRepo {
	return &ConfigRepo{DB: db}
}

// =====================
// Set (UPSERT)
// =====================
func (r *ConfigRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO config (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value
	`, key, value)

	return err
}

// =====================
// Get
// =====================
func (r *ConfigRepo) Get(ctx context.Context, key string) (*models.Config, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT key, value, created_at, updated_at
		FROM config
		WHERE key = ?
	`, key)

	var c models.Config
	err := row.Scan(
		&c.Key,
		&c.Value,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &c, err
}

// =====================
// GetValue (shortcut)
// =====================
func (r *ConfigRepo) GetValue(ctx context.Context, key string) (string, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT value FROM config WHERE key = ?
	`, key)

	var value string
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return value, nil
}
