CREATE TABLE IF NOT EXISTS manga (
  manga_id     INTEGER PRIMARY KEY AUTOINCREMENT,
  main_title   TEXT NOT NULL,
  description  TEXT,
  year         INTEGER,
  status_id    INTEGER,
  created_at   TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (status_id)
    REFERENCES manga_status(status_id)
    ON DELETE SET NULL
);
