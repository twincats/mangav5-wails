package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mangav5/internal/models"

	"github.com/wailsapp/wails/v3/pkg/application"
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

// GetMangaWithAlternativeTitles returns a manga with its alternative titles
func (r *MangaRepo) GetMangaWithAlternativeTitles(ctx context.Context, id int64) (*models.MangaWithAlt, error) {
	manga, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if manga == nil {
		return nil, nil
	}

	altTitles, err := r.GetAlternativeTitles(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.MangaWithAlt{
		Manga:             *manga,
		AlternativeTitles: altTitles,
	}, nil
}

// GetLatestManga returns the latest updated manga with their latest chapter
func (r *MangaRepo) GetLatestManga(ctx context.Context) ([]models.LatestManga, error) {
	query := `
		SELECT
		  m.manga_id        AS manga_id,
		  m.main_title      AS main_title,
		  ms.status_name    AS status_name,
		  c.chapter_id      AS chapter_id,
		  c.chapter_number  AS chapter_number,
		  c.created_at      AS download_time
		FROM manga AS m
		INNER JOIN chapters AS c
		  ON m.manga_id = c.manga_id
		INNER JOIN manga_status AS ms
		  ON m.status_id = ms.status_id
		WHERE
		  CAST(c.chapter_number AS REAL) = (
		    SELECT MAX(CAST(c2.chapter_number AS REAL))
		    FROM chapters AS c2
		    WHERE c2.manga_id = m.manga_id
		  )
		ORDER BY
		  c.created_at DESC,
		  m.main_title;
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.LatestManga
	for rows.Next() {
		var lm models.LatestManga
		if err := rows.Scan(
			&lm.MangaID,
			&lm.MainTitle,
			&lm.StatusName,
			&lm.ChapterID,
			&lm.ChapterNumber,
			&lm.DownloadTime,
		); err != nil {
			return nil, err
		}
		results = append(results, lm)
	}

	return results, nil
}

// ScanDirectoryForManga scans the given directory for manga and chapters
func (r *MangaRepo) ScanDirectoryForManga(ctx context.Context, mangasDir string) error {
	app := application.Get()
	entries, err := os.ReadDir(mangasDir)
	if err != nil {
		return err
	}

	// Count total manga directories for progress tracking
	totalManga := 0
	for _, e := range entries {
		if e.IsDir() {
			totalManga++
		}
	}

	chapterRepo := NewChapterRepo(r.DB)
	mangaIndex := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mangaIndex++

		mangaTitle := entry.Name()
		mangaPath := filepath.Join(mangasDir, mangaTitle)

		// Check if manga exists
		var mangaID int64
		err := r.DB.QueryRowContext(ctx, "SELECT manga_id FROM manga WHERE main_title = ?", mangaTitle).Scan(&mangaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Insert new manga
				newManga := &models.Manga{
					MainTitle:   mangaTitle,
					Description: fmt.Sprintf("Manga: %s", mangaTitle),
					Year:        time.Now().Year(),
					StatusID:    1, // Ongoing
				}
				mangaID, err = r.Insert(ctx, newManga)
				if err != nil {
					return fmt.Errorf("failed to insert manga %s: %w", mangaTitle, err)
				}
			} else {
				return fmt.Errorf("failed to query manga %s: %w", mangaTitle, err)
			}
		}

		// Scan for chapters
		subEntries, err := os.ReadDir(mangaPath)
		if err != nil {
			// Log error but continue with other mangas? Or return?
			// For now, let's return error to be safe
			return fmt.Errorf("failed to read directory %s: %w", mangaPath, err)
		}

		// Get existing chapters to avoid duplicates
		existingChapters, err := chapterRepo.GetByMangaID(ctx, mangaID)
		if err != nil {
			return fmt.Errorf("failed to get existing chapters for %s: %w", mangaTitle, err)
		}
		existingMap := make(map[float64]bool)
		for _, c := range existingChapters {
			existingMap[c.ChapterNumber] = true
		}

		var newChapters []models.Chapter
		for _, subEntry := range subEntries {
			name := subEntry.Name()

			// Determine if it's a chapter (dir or zip/cbz)
			isCompressed := 0
			ext := strings.ToLower(filepath.Ext(name))
			isArchive := ext == ".zip" || ext == ".cbz"

			if !subEntry.IsDir() && !isArchive {
				continue
			}

			if isArchive {
				isCompressed = 1
			}

			// Parse chapter number
			baseName := name
			if isArchive {
				baseName = strings.TrimSuffix(name, ext)
			}

			// Try parsing number
			chapterNum, err := parseChapterNumber(baseName)
			if err != nil {
				// Skip if not a valid chapter name
				continue
			}

			if existingMap[chapterNum] {
				continue
			}

			// Add to new chapters
			now := time.Now()
			newChapters = append(newChapters, models.Chapter{
				MangaID:         mangaID,
				ChapterNumber:   chapterNum,
				ChapterTitle:    fmt.Sprintf("Chapter %g", chapterNum),
				Volume:          0,
				TranslatorGroup: "Unknown",
				Language:        "en",
				ReleaseTimeTS:   now.Unix(),
				ReleaseTimeRaw:  now.Format("2006-01-02 15:04:05"),
				StatusRead:      0,
				Path:            filepath.ToSlash(filepath.Join(mangaTitle, name)), // Use forward slashes for portability
				IsCompressed:    isCompressed,
				Status:          "valid",
			})

			// Mark as existing in local map to handle duplicate files (e.g. folder and zip for same chapter)
			existingMap[chapterNum] = true
		}

		if len(newChapters) > 0 {
			if err := chapterRepo.BatchInsert(ctx, newChapters); err != nil {
				return fmt.Errorf("failed to batch insert chapters for %s: %w", mangaTitle, err)
			}
			// Emit event when chapter batch is finished
			if app != nil {
				app.Event.Emit("scanProgress", map[string]any{
					"mainTitle":     mangaTitle,
					"indexManga":    mangaIndex,
					"totalManga":    totalManga,
					"totalChapters": len(newChapters),
				})
			}
		}
	}

	return nil
}

// SaveManga inserts a new manga if it doesn't exist (by MainTitle), or retrieves the existing one.
// It sets default values for Description, Year, and StatusID if they are missing.
// Returns the manga ID and a boolean indicating if it was newly inserted (true) or retrieved (false).
func (r *MangaRepo) SaveManga(ctx context.Context, manga *models.Manga) (int64, bool, error) {
	// 1. If ID is already set, assume it's already saved/handled.
	if manga.ID != 0 {
		return manga.ID, false, nil
	}

	if manga.MainTitle == "" {
		return 0, false, errors.New("main_title is required")
	}

	// 2. Check if exists by MainTitle
	var existingID int64
	err := r.DB.QueryRowContext(ctx, "SELECT manga_id FROM manga WHERE main_title = ?", manga.MainTitle).Scan(&existingID)
	if err == nil {
		// Manga exists, update the ID in the struct
		manga.ID = existingID
		return existingID, false, nil
	} else if err != sql.ErrNoRows {
		return 0, false, fmt.Errorf("failed to check existing manga: %w", err)
	}

	// 3. Apply defaults
	if manga.Description == "" {
		manga.Description = fmt.Sprintf("Manga: %s", manga.MainTitle)
	}
	if manga.Year == 0 {
		manga.Year = time.Now().Year()
	}
	if manga.StatusID == 0 {
		manga.StatusID = 1 // Ongoing
	}

	// 4. Insert
	id, err := r.Insert(ctx, manga)
	if err != nil {
		return 0, false, fmt.Errorf("failed to insert manga: %w", err)
	}
	manga.ID = id

	// Emit event for new manga
	app := application.Get()
	if app != nil {
		app.Event.Emit("mangaSaved", manga)
	}

	return id, true, nil
}

// GetMangaDetail returns a manga with its alternative titles and chapters and manga_statis
func (r *MangaRepo) GetMangaDetail(ctx context.Context, id int64) (*models.MangaDetail, error) {
	manga, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if manga == nil {
		return nil, nil
	}

	var statusName string
	err = r.DB.QueryRowContext(ctx, "SELECT status_name FROM manga_status WHERE status_id = ?", manga.StatusID).Scan(&statusName)
	if err != nil {
		// If status not found, just ignore or set to unknown
		statusName = "Unknown"
	}

	altTitles, err := r.GetAlternativeTitles(ctx, id)
	if err != nil {
		return nil, err
	}

	chapterRepo := NewChapterRepo(r.DB)
	chapters, err := chapterRepo.GetByMangaID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.MangaDetail{
		Manga:             *manga,
		MangaStatus:       statusName,
		AlternativeTitles: altTitles,
		Chapters:          chapters,
	}, nil
}

// CheckStatusMangaByTitle check if manga title is in db or not
func (r *MangaRepo) CheckStatusMangaByTitle(ctx context.Context, title string) (bool, int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, "SELECT manga_id FROM manga WHERE main_title = ?", title).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

func parseChapterNumber(name string) (float64, error) {
	// Try direct parse
	if val, err := strconv.ParseFloat(name, 64); err == nil {
		return val, nil
	}

	// Try removing "Chapter " prefix (case insensitive)
	lowerName := strings.ToLower(name)
	if strings.HasPrefix(lowerName, "chapter") {
		clean := strings.TrimSpace(strings.TrimPrefix(lowerName, "chapter"))
		// Remove leading/trailing special chars just in case "Chapter - 10"
		clean = strings.Trim(clean, " -_")
		if val, err := strconv.ParseFloat(clean, 64); err == nil {
			return val, nil
		}
	}

	return 0, errors.New("invalid chapter number format")
}
