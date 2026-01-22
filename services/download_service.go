package services

import (
	"mangav5/downloader"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// DownloadService handles file downloading operations
type DownloadService struct{}

// NewDownloadService creates a new instance of DownloadService
func NewDownloadService() *DownloadService {
	return &DownloadService{}
}

// DownloadOptions allows configuring the download behavior
type DownloadOptions struct {
	MinConcurrency int `json:"minConcurrency"`
	MaxConcurrency int `json:"maxConcurrency"`
	RetryCount     int `json:"retryCount"`
	TimeoutSeconds int `json:"timeoutSeconds"`
}

// DownloadImages downloads a list of images to the specified output directory
// It uses an adaptive downloader engine to manage concurrency
func (s *DownloadService) DownloadImages(urls []string, outputDir string, options *DownloadOptions) error {
	// Default configuration
	cfg := downloader.DownloadConfig{
		MinConcurrency:   2,
		StartConcurrency: 4,
		MaxConcurrency:   8,
		RetryCount:       3,
		Timeout:          30 * time.Second,
		OutputDir:        outputDir,
	}

	// Apply options if provided
	if options != nil {
		if options.MinConcurrency > 0 {
			cfg.MinConcurrency = options.MinConcurrency
		}
		if options.MaxConcurrency > 0 {
			cfg.MaxConcurrency = options.MaxConcurrency
		}
		if options.RetryCount > 0 {
			cfg.RetryCount = options.RetryCount
		}
		if options.TimeoutSeconds > 0 {
			cfg.Timeout = time.Duration(options.TimeoutSeconds) * time.Second
		}
	}

	// Ensure StartConcurrency is within logical bounds
	if cfg.StartConcurrency > cfg.MaxConcurrency {
		cfg.StartConcurrency = cfg.MaxConcurrency
	}
	if cfg.StartConcurrency < cfg.MinConcurrency {
		cfg.StartConcurrency = cfg.MinConcurrency
	}

	// get application context
	app := application.Get()
	ctx := app.Context()

	// TODO: Add support for progress reporting callback or event emission
	// For now, we pass nil as the progress handler
	return downloader.DownloadImagesAdaptive(ctx, urls, cfg, func(report downloader.ProgressReport) {
		// Emit progress event to frontend
		app.Event.Emit("downloadProgress", report)
	})
}
