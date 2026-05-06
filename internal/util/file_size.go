package util

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

// DirSizeAuto menghitung total size directory secara otomatis
func DirSizeAuto(path string) int64 {
	const workerThreshold = 10000

	var totalSize int64
	fileCount := 0
	var fileCh chan fs.DirEntry
	var wg sync.WaitGroup

	startWorkers := func() {
		fileCh = make(chan fs.DirEntry, 1024)
		numWorkers := runtime.NumCPU() * 2
		if numWorkers < 1 {
			numWorkers = 1
		}
		if numWorkers > 32 {
			numWorkers = 32
		}
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for entry := range fileCh {
					info, err := entry.Info()
					if err != nil {
						continue
					}
					atomic.AddInt64(&totalSize, info.Size())
				}
			}()
		}
	}

	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			fileCount++
			if fileCh == nil && fileCount <= workerThreshold {
				info, err := d.Info()
				if err == nil {
					totalSize += info.Size()
				}
				return nil
			}
			if fileCh == nil {
				startWorkers()
			}
			fileCh <- d
		}
		return nil
	})
	if err != nil {
		log.Println("WalkDir error:", err)
	}

	if fileCh != nil {
		close(fileCh)
		wg.Wait()
	}

	return totalSize
}

// FormatFileSize sama seperti sebelumnya
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	switch exp {
	case 0:
		return fmt.Sprintf("%.2f KB", value)
	case 1:
		return fmt.Sprintf("%.2f MB", value)
	case 2:
		return fmt.Sprintf("%.2f GB", value)
	case 3:
		return fmt.Sprintf("%.2f TB", value)
	default:
		return fmt.Sprintf("%.2f PB", value)
	}
}
