package downloader

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
)

func downloadImage(
	ctx context.Context,
	client *resty.Client,
	url string,
	dir string,
	retry int,
) error {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var lastErr error

	for i := 0; i < retry; i++ {
		resp, err := client.R().SetContext(ctx).Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode() != 200 || len(resp.Body()) == 0 {
			lastErr = fmt.Errorf("bad response %d", resp.StatusCode())
			continue
		}

		ext := detectExt(resp)
		name := fmt.Sprintf("%x%s", sha1.Sum([]byte(url)), ext)
		path := filepath.Join(dir, name)

		if err := os.WriteFile(path, resp.Body(), 0644); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return lastErr
}
