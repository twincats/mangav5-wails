package util

import (
	"strconv"
	"strings"
	"time"
)

func ParseReleaseTime(raw string) (*int64, string) {
	raw = strings.TrimSpace(raw)

	// unix timestamp
	if ts, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return &ts, raw
	}

	layouts := []string{
		time.RFC3339,
		"02 Jan 2006 15:04",
		"January 2, 2006",
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			ts := t.UTC().Unix()
			return &ts, raw
		}
	}

	// relative time → skip for now
	return nil, raw
}
