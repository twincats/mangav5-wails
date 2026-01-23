CREATE TABLE IF NOT EXISTS chapters (
  chapter_id         INTEGER PRIMARY KEY AUTOINCREMENT,
  manga_id           INTEGER NOT NULL,
  chapter_number     REAL NOT NULL,
  chapter_title      TEXT,
  volume             INTEGER,
  translator_group   TEXT,
  language           TEXT,

  -- release time (RECOMMENDED)
  release_time_ts    INTEGER,        -- unix epoch (UTC)
  release_time_raw   TEXT,           -- raw scraped text

  status_read        INTEGER NOT NULL DEFAULT 0,
  path               TEXT,
  is_compressed      INTEGER NOT NULL DEFAULT 0,

  status             TEXT NOT NULL DEFAULT 'valid'
                     CHECK (status IN ('valid', 'missing', 'corrupted')),

  created_at         TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at         TEXT NOT NULL DEFAULT (datetime('now')),

  FOREIGN KEY (manga_id)
    REFERENCES manga(manga_id)
    ON DELETE CASCADE
);
