package services

import (
	"context"
	"mangav5/internal/downloader"
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

// DownloadImage downloads a single image to the specified output directory
func (s *DownloadService) DownloadImage(url string, outputDir string, baseName string, options *DownloadOptions) error {
	// Default configuration
	timeout := 30 * time.Second
	retry := 3

	// Apply options if provided
	if options != nil {
		if options.TimeoutSeconds > 0 {
			timeout = time.Duration(options.TimeoutSeconds) * time.Second
		}
		if options.RetryCount > 0 {
			retry = options.RetryCount
		}
	}

	// get application context
	ctx := context.Background()
	app := application.Get()
	if app != nil {
		ctx = app.Context()
	}

	client := downloader.NewRestyClient(timeout)

	return downloader.DownloadImage(ctx, client, url, outputDir, baseName, retry)
}
