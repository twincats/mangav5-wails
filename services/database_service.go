package services

import (
	"context"
	"errors"
	"mangav5/internal/models"
	"mangav5/internal/repo"
	"strings"
)

type DatabaseService struct {
	mangaRepo        *repo.MangaRepo
	configRepo       *repo.ConfigRepo
	chapterRepo      *repo.ChapterRepo
	scrapingRuleRepo *repo.ScrapingRuleRepo
}

func NewDatabaseService(repos *repo.Repositories) *DatabaseService {
	return &DatabaseService{
		mangaRepo:        repos.Manga,
		configRepo:       repos.Config,
		chapterRepo:      repos.Chapter,
		scrapingRuleRepo: repos.ScrapingRule,
	}
}

// =====================
// Manga Methods
// =====================

// CreateManga adds a new manga to the library with validation
func (s *DatabaseService) CreateManga(ctx context.Context, manga models.Manga) (int64, error) {
	// Business Logic: Validation
	if strings.TrimSpace(manga.MainTitle) == "" {
		return 0, errors.New("manga title cannot be empty")
	}

	// Business Logic: Default values
	if manga.Description == "" {
		manga.Description = "No description provided."
	}

	return s.mangaRepo.Insert(ctx, &manga)
}

func (s *DatabaseService) BatchCreateManga(ctx context.Context, mangas []models.Manga) error {
	return s.mangaRepo.BatchInsert(ctx, mangas)
}

// GetManga retrieves a manga by its ID
func (s *DatabaseService) GetManga(ctx context.Context, id int64) (*models.Manga, error) {
	return s.mangaRepo.GetByID(ctx, id)
}

func (s *DatabaseService) ListManga(ctx context.Context, limit, offset int) ([]models.Manga, error) {
	return s.mangaRepo.List(ctx, limit, offset)
}

func (s *DatabaseService) UpdateManga(ctx context.Context, manga models.Manga) error {
	return s.mangaRepo.Update(ctx, &manga)
}

func (s *DatabaseService) DeleteManga(ctx context.Context, id int64) error {
	return s.mangaRepo.Delete(ctx, id)
}

// AddAlternativeTitle adds an alternative title to a manga
func (s *DatabaseService) AddAlternativeTitle(ctx context.Context, mangaID int64, title string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("alternative title cannot be empty")
	}
	return s.mangaRepo.AddAlternativeTitle(ctx, mangaID, title)
}

// GetAlternativeTitles retrieves all alternative titles for a manga
func (s *DatabaseService) GetAlternativeTitles(ctx context.Context, mangaID int64) ([]models.AlternativeTitle, error) {
	return s.mangaRepo.GetAlternativeTitles(ctx, mangaID)
}

// GetAllMangaStatuses retrieves all available manga statuses
func (s *DatabaseService) GetAllMangaStatuses(ctx context.Context) ([]models.MangaStatus, error) {
	return s.mangaRepo.GetAllStatuses(ctx)
}

// GetMangaWithAlternativeTitles returns a manga with its alternative titles
func (s *DatabaseService) GetMangaWithAlternativeTitles(ctx context.Context, id int64) (*models.MangaWithAlt, error) {
	return s.mangaRepo.GetMangaWithAlternativeTitles(ctx, id)
}

// GetMangaDetail returns a manga with its alternative titles and chapters
func (s *DatabaseService) GetMangaDetail(ctx context.Context, id int64) (*models.MangaDetail, error) {
	return s.mangaRepo.GetMangaDetail(ctx, id)
}

// GetLatestManga returns the latest updated manga
func (s *DatabaseService) GetLatestManga(ctx context.Context) ([]models.LatestManga, error) {
	return s.mangaRepo.GetLatestManga(ctx)
}

// ScanDirectoryForManga scans the given directory for manga and chapters
func (s *DatabaseService) ScanDirectoryForManga(ctx context.Context, mangasDir string) error {
	if strings.TrimSpace(mangasDir) == "" {
		return errors.New("directory path cannot be empty")
	}
	return s.mangaRepo.ScanDirectoryForManga(ctx, mangasDir)
}

// SaveManga inserts a new manga if it doesn't exist, or retrieves the existing one.
// Returns the manga ID and a boolean indicating if it was newly inserted (true) or retrieved (false).
func (s *DatabaseService) SaveManga(ctx context.Context, manga models.Manga) (int64, bool, error) {
	return s.mangaRepo.SaveManga(ctx, &manga)
}

// CheckStatusMangaByTitle checks if a manga with the given title exists.
func (s *DatabaseService) CheckStatusMangaByTitle(ctx context.Context, title string) (bool, int64, error) {
	return s.mangaRepo.CheckStatusMangaByTitle(ctx, title)
}

// =====================
// Chapter Methods
// =====================

func (s *DatabaseService) CreateChapter(ctx context.Context, chapter models.Chapter) (int64, error) {
	return s.chapterRepo.Insert(ctx, &chapter)
}

func (s *DatabaseService) BatchCreateChapters(ctx context.Context, chapters []models.Chapter) error {
	return s.chapterRepo.BatchInsert(ctx, chapters)
}

func (s *DatabaseService) GetChaptersByMangaID(ctx context.Context, mangaID int64) ([]models.Chapter, error) {
	return s.chapterRepo.GetByMangaID(ctx, mangaID)
}

func (s *DatabaseService) GetChapter(ctx context.Context, id int64) (*models.Chapter, error) {
	return s.chapterRepo.GetByID(ctx, id)
}

func (s *DatabaseService) UpdateChapter(ctx context.Context, chapter models.Chapter) error {
	return s.chapterRepo.Update(ctx, &chapter)
}

func (s *DatabaseService) DeleteChapter(ctx context.Context, id int64) error {
	return s.chapterRepo.Delete(ctx, id)
}

// MarkChapterAsRead updates the chapter status_read to true
func (s *DatabaseService) MarkChapterAsRead(ctx context.Context, chapterID int64) error {
	return s.chapterRepo.UpdateStatusRead(ctx, chapterID)
}

// =====================
// Scraping Rule Methods
// =====================

func (s *DatabaseService) SaveScrapingRule(ctx context.Context, rule models.ScrapingRule) error {
	return s.scrapingRuleRepo.Upsert(ctx, &rule)
}

func (s *DatabaseService) ListScrapingRules(ctx context.Context) ([]models.ScrapingRule, error) {
	return s.scrapingRuleRepo.List(ctx)
}

func (s *DatabaseService) ListScrapingRulesBasic(ctx context.Context) ([]models.ScrapingRule, error) {
	return s.scrapingRuleRepo.ListBasic(ctx)
}

func (s *DatabaseService) GetScrapingRule(ctx context.Context, siteKey string) (*models.ScrapingRule, error) {
	return s.scrapingRuleRepo.GetBySiteKey(ctx, siteKey)
}

func (s *DatabaseService) DeleteScrapingRule(ctx context.Context, siteKey string) error {
	return s.scrapingRuleRepo.Delete(ctx, siteKey)
}

// =====================
// Config Methods
// =====================

// SetConfig sets a configuration value for a given key
func (s *DatabaseService) SetConfig(ctx context.Context, key, value string) error {
	return s.configRepo.Set(ctx, key, value)
}

// GetConfig retrieves a configuration object by key
func (s *DatabaseService) GetConfig(ctx context.Context, key string) (*models.Config, error) {
	return s.configRepo.Get(ctx, key)
}

// GetConfigValue retrieves just the value string for a given key
func (s *DatabaseService) GetConfigValue(ctx context.Context, key string) (string, error) {
	return s.configRepo.GetValue(ctx, key)
}
