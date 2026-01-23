CREATE INDEX IF NOT EXISTS idx_chapters_manga
ON chapters(manga_id);

CREATE INDEX IF NOT EXISTS idx_chapters_release_ts
ON chapters(release_time_ts);
