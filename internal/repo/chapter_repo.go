package repo

import (
	"context"
	"database/sql"
	"errors"

	"mangav5/internal/models"
	"mangav5/internal/util"
)

type ChapterRepo struct {
	DB *sql.DB
}

func NewChapterRepo(db *sql.DB) *ChapterRepo {
	return &ChapterRepo{DB: db}
}

// Insert
func (r *ChapterRepo) Insert(ctx context.Context, c *models.Chapter) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `
		INSERT INTO chapters (
			manga_id, chapter_number, chapter_title, volume, translator_group, language,
			release_time_ts, release_time_raw, status_read, path, is_compressed, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.MangaID, c.ChapterNumber, c.ChapterTitle, c.Volume, c.TranslatorGroup, c.Language,
		c.ReleaseTimeTS, c.ReleaseTimeRaw, c.StatusRead, c.Path, c.IsCompressed, c.Status)

	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Batch Insert
func (r *ChapterRepo) BatchInsert(ctx context.Context, chapters []models.Chapter) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chapters (
			manga_id, chapter_number, chapter_title, volume, translator_group, language,
			release_time_ts, release_time_raw, status_read, path, is_compressed, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range chapters {
		if c.ReleaseTimeTS == 0 && c.ReleaseTimeRaw != "" {
			if ts, _ := util.ParseReleaseTime(c.ReleaseTimeRaw); ts != nil {
				c.ReleaseTimeTS = *ts
			}
		}
		_, err := stmt.ExecContext(ctx,
			c.MangaID, c.ChapterNumber, c.ChapterTitle, c.Volume, c.TranslatorGroup, c.Language,
			c.ReleaseTimeTS, c.ReleaseTimeRaw, c.StatusRead, c.Path, c.IsCompressed, c.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByMangaID
func (r *ChapterRepo) GetByMangaID(ctx context.Context, mangaID int64) ([]models.Chapter, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT 
			chapter_id, manga_id, chapter_number, chapter_title, volume, translator_group, language,
			release_time_ts, release_time_raw, status_read, path, is_compressed, status,
			created_at, updated_at
		FROM chapters
		WHERE manga_id = ?
		ORDER BY chapter_number DESC
	`, mangaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Chapter
	for rows.Next() {
		var c models.Chapter
		var releaseTimeTS sql.NullInt64
		var releaseTimeRaw, chapterTitle, translatorGroup, language, path, status sql.NullString
		var volume sql.NullInt64

		if err := rows.Scan(
			&c.ID, &c.MangaID, &c.ChapterNumber, &chapterTitle, &volume, &translatorGroup, &language,
			&releaseTimeTS, &releaseTimeRaw, &c.StatusRead, &path, &c.IsCompressed, &status,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.ChapterTitle = chapterTitle.String
		c.Volume = int(volume.Int64)
		c.TranslatorGroup = translatorGroup.String
		c.Language = language.String
		c.ReleaseTimeTS = releaseTimeTS.Int64
		c.ReleaseTimeRaw = releaseTimeRaw.String
		c.Path = path.String
		c.Status = status.String
		result = append(result, c)
	}
	return result, nil
}

// GetByID
func (r *ChapterRepo) GetByID(ctx context.Context, id int64) (*models.Chapter, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT 
			chapter_id, manga_id, chapter_number, chapter_title, volume, translator_group, language,
			release_time_ts, release_time_raw, status_read, path, is_compressed, status,
			created_at, updated_at
		FROM chapters
		WHERE chapter_id = ?
	`, id)

	var c models.Chapter
	var releaseTimeTS sql.NullInt64
	var releaseTimeRaw, chapterTitle, translatorGroup, language, path, status sql.NullString
	var volume sql.NullInt64

	err := row.Scan(
		&c.ID, &c.MangaID, &c.ChapterNumber, &chapterTitle, &volume, &translatorGroup, &language,
		&releaseTimeTS, &releaseTimeRaw, &c.StatusRead, &path, &c.IsCompressed, &status,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.ChapterTitle = chapterTitle.String
	c.Volume = int(volume.Int64)
	c.TranslatorGroup = translatorGroup.String
	c.Language = language.String
	c.ReleaseTimeTS = releaseTimeTS.Int64
	c.ReleaseTimeRaw = releaseTimeRaw.String
	c.Path = path.String
	c.Status = status.String

	return &c, nil
}

// Update
func (r *ChapterRepo) Update(ctx context.Context, c *models.Chapter) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE chapters
		SET chapter_number=?, chapter_title=?, volume=?, translator_group=?, language=?,
			release_time_ts=?, release_time_raw=?, status_read=?, path=?, is_compressed=?, status=?, updated_at=datetime('now')
		WHERE chapter_id=?
	`, c.ChapterNumber, c.ChapterTitle, c.Volume, c.TranslatorGroup, c.Language,
		c.ReleaseTimeTS, c.ReleaseTimeRaw, c.StatusRead, c.Path, c.IsCompressed, c.Status, c.ID)
	return err
}

// Delete
func (r *ChapterRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM chapters WHERE chapter_id=?`, id)
	return err
}
