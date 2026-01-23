-- trigger config
CREATE TRIGGER trg_config_updated
AFTER UPDATE ON config
FOR EACH ROW
BEGIN
  UPDATE config
  SET updated_at = datetime('now')
  WHERE config_id = OLD.config_id;
END;

-- trigger manga
CREATE TRIGGER trg_manga_updated
AFTER UPDATE ON manga
FOR EACH ROW
BEGIN
  UPDATE manga
  SET updated_at = datetime('now')
  WHERE manga_id = OLD.manga_id;
END;

-- trigger chapters
CREATE TRIGGER trg_chapters_updated
AFTER UPDATE ON chapters
FOR EACH ROW
BEGIN
  UPDATE chapters
  SET updated_at = datetime('now')
  WHERE chapter_id = OLD.chapter_id;
END;

-- trigger scraping_rules
CREATE TRIGGER trg_scraping_rules_updated
AFTER UPDATE ON scraping_rules
FOR EACH ROW
BEGIN
  UPDATE scraping_rules
  SET updated_at = datetime('now')
  WHERE id = OLD.id;
END;
