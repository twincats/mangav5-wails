package downloader

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

func DownloadImagesAdaptive(
	ctx context.Context,
	urls []string,
	cfg DownloadConfig,
	onProgress func(ProgressReport),
) error {

	client := NewRestyClient(cfg.Timeout)
	ctrl := NewAdaptiveController(
		cfg.StartConcurrency,
		cfg.MinConcurrency,
		cfg.MaxConcurrency,
	)

	total := len(urls)
	jobs := make(chan int)
	results := make(chan struct {
		index   int
		success bool
	}, total)

	// semaphore controls concurrency
	sem := make(chan struct{}, cfg.MaxConcurrency)

	// Calculate padding width for filenames
	padWidth := len(fmt.Sprintf("%d", total))

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()

		for idx := range jobs {
			// acquire token
			sem <- struct{}{}

			url := urls[idx]

			// Generate filename based on index (1-based) with padding
			baseName := fmt.Sprintf("%0*d", padWidth, idx+1)

			start := time.Now()
			err := downloadImage(ctx, client, url, cfg.OutputDir, baseName, cfg.RetryCount)
			latency := time.Since(start)

			// adaptive logic
			ctrl.AddLatency(latency)
			ctrl.Adjust(err == nil)

			// release token
			<-sem

			results <- struct {
				index   int
				success bool
			}{idx, err == nil}
		}
	}

	// start fixed number of workers (MAX)
	wg.Add(cfg.MaxConcurrency)
	for i := 0; i < cfg.MaxConcurrency; i++ {
		go worker()
	}

	// feed jobs
	go func() {
		for i := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}
		}
		close(jobs)
	}()

	// progress dispatcher
	completed := 0
	for completed < total {
		select {
		case res := <-results:
			completed++

			status := "success"
			if !res.success {
				status = "fail"
			}

			if onProgress != nil {
				onProgress(ProgressReport{
					Index:    completed,
					Total:    total,
					Filename: filepath.Base(urls[res.index]),
					Status:   status,
				})
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	wg.Wait()
	return nil
}

// sample config
// cfg := DownloadConfig{
// 	MinConcurrency:    2,
// 	StartConcurrency:  4,
// 	MaxConcurrency:    8,
// 	RetryCount:        3,
// 	Timeout:           15 * time.Second,
// 	OutputDir:         "./images",
// }
