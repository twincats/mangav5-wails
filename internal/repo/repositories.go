package repo

import "database/sql"

type Repositories struct {
	Manga        *MangaRepo
	Config       *ConfigRepo
	Chapter      *ChapterRepo
	ScrapingRule *ScrapingRuleRepo
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Manga:        NewMangaRepo(db),
		Config:       NewConfigRepo(db),
		Chapter:      NewChapterRepo(db),
		ScrapingRule: NewScrapingRuleRepo(db),
	}
}
