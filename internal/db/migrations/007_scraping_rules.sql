CREATE TABLE IF NOT EXISTS scraping_rules (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  site_key           TEXT NOT NULL UNIQUE,
  name               TEXT NOT NULL,
  domains_json       TEXT NOT NULL,   -- JSON array of domains
  website_url        TEXT NOT NULL,

  manga_rule_json    TEXT NOT NULL,
  chapter_rule_json  TEXT NOT NULL,

  enabled            INTEGER NOT NULL DEFAULT 1,
  priority           INTEGER NOT NULL DEFAULT 100,

  created_at         TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at         TEXT NOT NULL DEFAULT (datetime('now'))
);
