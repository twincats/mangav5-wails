package downloader

import (
	"time"
)

type DownloadConfig struct {
	MinConcurrency   int
	StartConcurrency int
	MaxConcurrency   int

	RetryCount int
	Timeout    time.Duration
	OutputDir  string
}

type ProgressReport struct {
	Index    int    `json:"index"`
	Total    int    `json:"total"`
	Filename string `json:"filename"`
	Status   string `json:"status"` // success | fail
}
