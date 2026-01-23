package models

type Chapter struct {
	ID              int64   `json:"id"`
	MangaID         int64   `json:"manga_id"`
	ChapterNumber   float64 `json:"chapter_number"`
	ChapterTitle    string  `json:"chapter_title"`
	Volume          int     `json:"volume"`
	TranslatorGroup string  `json:"translator_group"`
	Language        string  `json:"language"`
	ReleaseTimeTS   int64   `json:"release_time_ts"`
	ReleaseTimeRaw  string  `json:"release_time_raw"`
	StatusRead      int     `json:"status_read"` // 0 or 1
	Path            string  `json:"path"`
	IsCompressed    int     `json:"is_compressed"` // 0 or 1
	Status          string  `json:"status"`        // valid, missing, corrupted
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
