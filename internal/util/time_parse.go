package util

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Regex for relative time
	reHoursAgo   = regexp.MustCompile(`(?i)(\d+)\s*(?:jam|hour|hours)\s*(?:yang lalu|ago)`)
	reMinutesAgo = regexp.MustCompile(`(?i)(\d+)\s*(?:menit|minute|minutes)\s*(?:yang lalu|ago)`)
	reDaysAgo    = regexp.MustCompile(`(?i)(\d+)\s*(?:hari|day|days)\s*(?:yang lalu|ago)`)
	reYearsAgo   = regexp.MustCompile(`(?i)(\d+)\s*(?:tahun|year|years)\s*(?:yang lalu|ago)`)
	reYesterday  = regexp.MustCompile(`(?i)\b(?:kemarin|yesterday)\b`)
)

func ParseReleaseTime(raw string) (*int64, string) {
	raw = strings.TrimSpace(raw)

	// unix timestamp
	if ts, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return &ts, raw
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",

		// YYYY-MM-DD
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",

		// YYYY/MM/DD
		"2006/01/02 15:04:05",
		"2006/01/02",

		// Explicit Month (English) - Unambiguous
		"02 Jan 2006 15:04",
		"02 Jan 2006",
		"2 Jan 2006",
		"Jan 02, 2006",
		"Jan 2, 2006",
		"January 02, 2006",
		"January 2, 2006",

		// DD/MM/YYYY (Common in ID/EU)
		"02/01/2006",
		"02-01-2006",
		"02.01.2006",

		// US Format
		"01/02/2006",
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			ts := t.UTC().Unix()
			return &ts, raw
		}
	}

	// Relative time parsing
	now := time.Now().UTC()

	// Hours ago / Jam yang lalu
	if matches := reHoursAgo.FindStringSubmatch(raw); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			ts := now.Add(-time.Duration(val) * time.Hour).Unix()
			return &ts, raw
		}
	}

	// Minutes ago / Menit yang lalu
	if matches := reMinutesAgo.FindStringSubmatch(raw); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			ts := now.Add(-time.Duration(val) * time.Minute).Unix()
			return &ts, raw
		}
	}

	// Days ago / Hari yang lalu
	if matches := reDaysAgo.FindStringSubmatch(raw); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			ts := now.AddDate(0, 0, -val).Unix()
			return &ts, raw
		}
	}

	// Years ago / Tahun yang lalu
	if matches := reYearsAgo.FindStringSubmatch(raw); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			ts := now.AddDate(-val, 0, 0).Unix()
			return &ts, raw
		}
	}

	// Yesterday / Kemarin
	if reYesterday.MatchString(raw) {
		ts := now.AddDate(0, 0, -1).Unix()
		return &ts, raw
	}

	// relative time → skip for now
	return nil, raw
}
