CREATE TABLE IF NOT EXISTS alternative_titles (
  alt_id            INTEGER PRIMARY KEY AUTOINCREMENT,
  manga_id          INTEGER NOT NULL,
  alternative_title TEXT NOT NULL,
  created_at        TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (manga_id)
    REFERENCES manga(manga_id)
    ON DELETE CASCADE,
  UNIQUE (manga_id, alternative_title)
);
