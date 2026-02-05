package models

type Manga struct {
	ID          int64  `json:"id"`
	MainTitle   string `json:"main_title"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	StatusID    int64  `json:"status_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type AlternativeTitle struct {
	ID               int64  `json:"id"`
	MangaID          int64  `json:"manga_id"`
	AlternativeTitle string `json:"alternative_title"`
	CreatedAt        string `json:"created_at"`
}

type MangaStatus struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type MangaWithAlt struct {
	Manga
	AlternativeTitles []AlternativeTitle `json:"alternative_titles"`
}

type MangaDetail struct {
	Manga
	MangaStatus       string             `json:"manga_status"`
	AlternativeTitles []AlternativeTitle `json:"alternative_titles"`
	Chapters          []Chapter          `json:"chapters"`
}

type LatestManga struct {
	MangaID       int64   `json:"manga_id"`
	MainTitle     string  `json:"main_title"`
	StatusName    string  `json:"status_name"`
	ChapterID     int64   `json:"chapter_id"`
	ChapterNumber float64 `json:"chapter_number"`
	DownloadTime  string  `json:"download_time"`
}
