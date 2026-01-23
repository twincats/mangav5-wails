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

func NewDatabaseService(
	mangaRepo *repo.MangaRepo,
	configRepo *repo.ConfigRepo,
	chapterRepo *repo.ChapterRepo,
	scrapingRuleRepo *repo.ScrapingRuleRepo,
) *DatabaseService {
	return &DatabaseService{
		mangaRepo:        mangaRepo,
		configRepo:       configRepo,
		chapterRepo:      chapterRepo,
		scrapingRuleRepo: scrapingRuleRepo,
	}
}

// =====================
// Manga Methods
// =====================

// CreateManga adds a new manga to the library with validation
func (s *DatabaseService) CreateManga(manga models.Manga) (int64, error) {
	// Business Logic: Validation
	if strings.TrimSpace(manga.MainTitle) == "" {
		return 0, errors.New("manga title cannot be empty")
	}

	// Business Logic: Default values
	if manga.Description == "" {
		manga.Description = "No description provided."
	}

	return s.mangaRepo.Insert(context.Background(), &manga)
}

// GetManga retrieves a manga by its ID
func (s *DatabaseService) GetManga(id int64) (*models.Manga, error) {
	return s.mangaRepo.GetByID(context.Background(), id)
}

// AddAlternativeTitle adds an alternative title to a manga
func (s *DatabaseService) AddAlternativeTitle(mangaID int64, title string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("alternative title cannot be empty")
	}
	return s.mangaRepo.AddAlternativeTitle(context.Background(), mangaID, title)
}

// GetAlternativeTitles retrieves all alternative titles for a manga
func (s *DatabaseService) GetAlternativeTitles(mangaID int64) ([]models.AlternativeTitle, error) {
	return s.mangaRepo.GetAlternativeTitles(context.Background(), mangaID)
}

// GetAllMangaStatuses retrieves all available manga statuses
func (s *DatabaseService) GetAllMangaStatuses() ([]models.MangaStatus, error) {
	return s.mangaRepo.GetAllStatuses(context.Background())
}

// =====================
// Chapter Methods
// =====================

func (s *DatabaseService) CreateChapter(chapter models.Chapter) (int64, error) {
	return s.chapterRepo.Insert(context.Background(), &chapter)
}

func (s *DatabaseService) BatchCreateChapters(chapters []models.Chapter) error {
	return s.chapterRepo.BatchInsert(context.Background(), chapters)
}

func (s *DatabaseService) GetChaptersByMangaID(mangaID int64) ([]models.Chapter, error) {
	return s.chapterRepo.GetByMangaID(context.Background(), mangaID)
}

// =====================
// Scraping Rule Methods
// =====================

func (s *DatabaseService) ListScrapingRules() ([]models.ScrapingRule, error) {
	return s.scrapingRuleRepo.List(context.Background())
}

func (s *DatabaseService) GetScrapingRule(siteKey string) (*models.ScrapingRule, error) {
	return s.scrapingRuleRepo.GetBySiteKey(context.Background(), siteKey)
}

func (s *DatabaseService) UpdateScrapingRule(rule models.ScrapingRule) error {
	return s.scrapingRuleRepo.Update(context.Background(), &rule)
}

// =====================
// Config Methods
// =====================

// SetConfig sets a configuration value for a given key
func (s *DatabaseService) SetConfig(key, value string) error {
	return s.configRepo.Set(context.Background(), key, value)
}

// GetConfig retrieves a configuration object by key
func (s *DatabaseService) GetConfig(key string) (*models.Config, error) {
	return s.configRepo.Get(context.Background(), key)
}

// GetConfigValue retrieves just the value string for a given key
func (s *DatabaseService) GetConfigValue(key string) (string, error) {
	return s.configRepo.GetValue(context.Background(), key)
}
