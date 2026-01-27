package repo

import (
	"context"
	"database/sql"
	"errors"

	"mangav5/internal/models"
)

type ScrapingRuleRepo struct {
	DB *sql.DB
}

func NewScrapingRuleRepo(db *sql.DB) *ScrapingRuleRepo {
	return &ScrapingRuleRepo{DB: db}
}

func (r *ScrapingRuleRepo) Insert(ctx context.Context, rule *models.ScrapingRule) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `
		INSERT INTO scraping_rules (site_key, name, domains_json, manga_rule_json, chapter_rule_json, enabled)
		VALUES (?, ?, ?, ?, ?, ?)
	`, rule.SiteKey, rule.Name, rule.DomainsJSON, rule.MangaRuleJSON, rule.ChapterRuleJSON, rule.Enabled)

	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ScrapingRuleRepo) Upsert(ctx context.Context, rule *models.ScrapingRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO scraping_rules (site_key, name, domains_json, manga_rule_json, chapter_rule_json, enabled)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(site_key) DO UPDATE SET
			name = excluded.name,
			domains_json = excluded.domains_json,
			manga_rule_json = excluded.manga_rule_json,
			chapter_rule_json = excluded.chapter_rule_json,
			enabled = excluded.enabled,
			updated_at = datetime('now')
	`, rule.SiteKey, rule.Name, rule.DomainsJSON, rule.MangaRuleJSON, rule.ChapterRuleJSON, rule.Enabled)

	return err
}

func (r *ScrapingRuleRepo) List(ctx context.Context) ([]models.ScrapingRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, site_key, name, domains_json, manga_rule_json, chapter_rule_json, enabled, created_at, updated_at
		FROM scraping_rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ScrapingRule
	for rows.Next() {
		var s models.ScrapingRule
		if err := rows.Scan(
			&s.ID, &s.SiteKey, &s.Name, &s.DomainsJSON,
			&s.MangaRuleJSON, &s.ChapterRuleJSON, &s.Enabled,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *ScrapingRuleRepo) GetBySiteKey(ctx context.Context, siteKey string) (*models.ScrapingRule, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, site_key, name, domains_json, manga_rule_json, chapter_rule_json, enabled, created_at, updated_at
		FROM scraping_rules
		WHERE site_key = ?
	`, siteKey)

	var s models.ScrapingRule
	err := row.Scan(
		&s.ID, &s.SiteKey, &s.Name, &s.DomainsJSON,
		&s.MangaRuleJSON, &s.ChapterRuleJSON, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ScrapingRuleRepo) Update(ctx context.Context, rule *models.ScrapingRule) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE scraping_rules
		SET name=?, domains_json=?, manga_rule_json=?, chapter_rule_json=?, enabled=?, updated_at=datetime('now')
		WHERE site_key=?
	`, rule.Name, rule.DomainsJSON, rule.MangaRuleJSON, rule.ChapterRuleJSON, rule.Enabled, rule.SiteKey)
	return err
}

func (r *ScrapingRuleRepo) Delete(ctx context.Context, siteKey string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM scraping_rules WHERE site_key=?`, siteKey)
	return err
}

func (r *ScrapingRuleRepo) ListBasic(ctx context.Context) ([]models.ScrapingRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, site_key, name, domains_json
		FROM scraping_rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ScrapingRule
	for rows.Next() {
		var s models.ScrapingRule
		if err := rows.Scan(&s.ID, &s.SiteKey, &s.Name, &s.DomainsJSON); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}
