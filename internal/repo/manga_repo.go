package repo

import (
	"context"
	"database/sql"
	"errors"

	"mangav5/internal/models"
)

type MangaRepo struct {
	DB *sql.DB
}

func NewMangaRepo(db *sql.DB) *MangaRepo {
	return &MangaRepo{DB: db}
}

// =====================
// Insert
// =====================
func (r *MangaRepo) Insert(ctx context.Context, manga *models.Manga) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `
		INSERT INTO manga (main_title, description, year, status_id)
		VALUES (?, ?, ?, ?)
	`, manga.MainTitle, manga.Description, manga.Year, manga.StatusID)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// =====================
// Batch Insert
// =====================
func (r *MangaRepo) BatchInsert(ctx context.Context, mangas []models.Manga) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO manga (main_title, description, year, status_id)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range mangas {
		if _, err := stmt.ExecContext(ctx, m.MainTitle, m.Description, m.Year, m.StatusID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// =====================
// Get by ID
// =====================
func (r *MangaRepo) GetByID(ctx context.Context, id int64) (*models.Manga, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT manga_id, main_title, description, year, status_id, created_at, updated_at
		FROM manga
		WHERE manga_id = ?
	`, id)

	var m models.Manga
	err := row.Scan(
		&m.ID,
		&m.MainTitle,
		&m.Description,
		&m.Year,
		&m.StatusID,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &m, err
}

// =====================
// List (pagination ready)
// =====================
func (r *MangaRepo) List(ctx context.Context, limit, offset int) ([]models.Manga, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT manga_id, main_title, description, year, status_id, created_at, updated_at
		FROM manga
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Manga
	for rows.Next() {
		var m models.Manga
		if err := rows.Scan(
			&m.ID,
			&m.MainTitle,
			&m.Description,
			&m.Year,
			&m.StatusID,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}

// =====================
// Update
// =====================
func (r *MangaRepo) Update(ctx context.Context, manga *models.Manga) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE manga
		SET main_title = ?, description = ?, year = ?, status_id = ?, updated_at = datetime('now')
		WHERE manga_id = ?
	`, manga.MainTitle, manga.Description, manga.Year, manga.StatusID, manga.ID)

	return err
}

// =====================
// Alternative Titles
// =====================
func (r *MangaRepo) AddAlternativeTitle(ctx context.Context, mangaID int64, title string) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO alternative_titles (manga_id, alternative_title)
		VALUES (?, ?)
		ON CONFLICT(manga_id, alternative_title) DO NOTHING
	`, mangaID, title)
	return err
}

func (r *MangaRepo) GetAlternativeTitles(ctx context.Context, mangaID int64) ([]models.AlternativeTitle, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT alt_id, manga_id, alternative_title, created_at
		FROM alternative_titles
		WHERE manga_id = ?
	`, mangaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AlternativeTitle
	for rows.Next() {
		var t models.AlternativeTitle
		if err := rows.Scan(&t.ID, &t.MangaID, &t.AlternativeTitle, &t.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

// =====================
// Manga Status
// =====================
func (r *MangaRepo) GetAllStatuses(ctx context.Context) ([]models.MangaStatus, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT status_id, status_name FROM manga_status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.MangaStatus
	for rows.Next() {
		var s models.MangaStatus
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

// =====================
// Delete
// =====================
func (r *MangaRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `
		DELETE FROM manga WHERE id = ?
	`, id)
	return err
}
