package models

type ScrapingRule struct {
	ID              int64  `json:"id"`
	SiteKey         string `json:"site_key"`
	Name            string `json:"name"`
	DomainsJSON     string `json:"domains_json"`
	WebsiteURL      string `json:"website_url"`
	MangaRuleJSON   string `json:"manga_rule_json"`
	ChapterRuleJSON string `json:"chapter_rule_json"`
	Enabled         int    `json:"enabled"`
	Priority        int    `json:"priority"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
