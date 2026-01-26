package util

import (
	"testing"
	"time"
)

func TestParseReleaseTime(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantTime bool // if true, we check if timestamp is reasonably close to expected relative time
		relDiff  time.Duration
	}{
		{"unix timestamp", "1700000000", true, 0}, // special handling in test logic
		{"RFC3339", "2023-10-27T10:00:00Z", true, 0},
		{"Date Only", "2023-10-27", true, 0},
		{"Hours Ago", "2 hours ago", true, -2 * time.Hour},
		{"Jam Yang Lalu", "5 jam yang lalu", true, -5 * time.Hour},
		{"Minutes Ago", "30 minutes ago", true, -30 * time.Minute},
		{"Menit Yang Lalu", "10 menit yang lalu", true, -10 * time.Minute},
		{"Days Ago", "2 days ago", true, -48 * time.Hour},
		{"Yesterday", "Yesterday", true, -24 * time.Hour},
		{"Kemarin", "kemarin", true, -24 * time.Hour},
		{"Years Ago", "1 year ago", true, -24 * 365 * time.Hour}, // approximate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseReleaseTime(tt.raw)
			if got == nil {
				t.Errorf("ParseReleaseTime(%q) returned nil", tt.raw)
				return
			}

			// For relative tests, we check if the result is close to time.Now().Add(diff)
			if tt.relDiff != 0 {
				expected := time.Now().UTC().Add(tt.relDiff).Unix()
				diff := *got - expected
				if diff < -5 || diff > 5 { // Allow 5 seconds margin
					t.Errorf("ParseReleaseTime(%q) = %v, want close to %v (diff %d seconds)", tt.raw, *got, expected, diff)
				}
			} else if tt.name == "unix timestamp" {
				if *got != 1700000000 {
					t.Errorf("ParseReleaseTime(%q) = %v, want 1700000000", tt.raw, *got)
				}
			}
		})
	}
}
