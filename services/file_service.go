package services

import (
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mangav5/internal/zipper"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// FileService handles file operations for manga
type FileService struct {
	dbService *DatabaseService
}

var MANGA_DIR = ""

// NewFileService creates a new FileService
func NewFileService(dbService *DatabaseService) *FileService {
	return &FileService{dbService: dbService}
}

func (s *FileService) GetMangaDir() (string, error) {
	if MANGA_DIR == "" {
		ctx := application.Get().Context()
		mangaDir, err := s.dbService.GetConfig(ctx, "manga_directory")
		if err != nil {
			return "", err
		}
		if mangaDir == nil {
			return "", errors.New("manga directory not configured")
		}
		MANGA_DIR = mangaDir.Value
		return mangaDir.Value, nil
	}
	return MANGA_DIR, nil
}

func (s *FileService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mangaDir, err := s.GetMangaDir()
	if err != nil {
		http.Error(w, "failed to get manga directory", http.StatusInternalServerError)
		return
	}

	// Remove the route prefix "/filemanga" if present so we get the relative file path
	requestPath := strings.TrimPrefix(r.URL.Path, "/filemanga")

	// Construct full file path
	fullPath := filepath.Join(mangaDir, requestPath)

	// Security: Prevent directory traversal
	if !strings.HasPrefix(fullPath, mangaDir) {
		http.Error(w, "invalid file path", http.StatusForbidden)
		return
	}

	// Check if file exists; if not, check for archive
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Attempt to serve from zip/cbz archive
		dir := filepath.Dir(fullPath)
		filename := filepath.Base(fullPath)

		// Check for .cbz or .zip corresponding to the directory
		var archivePath string
		cbzPath := dir + ".cbz"
		zipPath := dir + ".zip"

		if _, err := os.Stat(cbzPath); err == nil {
			archivePath = cbzPath
		} else if _, err := os.Stat(zipPath); err == nil {
			archivePath = zipPath
		}

		if archivePath != "" {
			// Try to read the file from the archive
			rc, err := zipper.OpenFileFromArchive(archivePath, filename)
			if err == nil {
				defer rc.Close()
				w.Header().Set("Cache-Control", "public, max-age=3600")

				mimeType := mime.TypeByExtension(filepath.Ext(filename))
				if mimeType == "" {
					mimeType = "application/octet-stream"
				}
				w.Header().Set("Content-Type", mimeType)

				// Use io.Copy to stream data directly to the response writer
				// This avoids loading the entire file into memory
				if _, err := io.Copy(w, rc); err != nil {
					// Handle error (optional log)
				}
				return
			}
		}
	}

	// Fallback logic for cover.webp or cover
	// If cover.webp is requested but does not exist, try:
	// 1. "cover" with other extensions in the same directory
	// 2. Any first image in the same directory
	// 3. Any first image in subdirectories
	baseName := filepath.Base(fullPath)
	if strings.EqualFold(baseName, "cover.webp") || strings.EqualFold(baseName, "cover") {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			dir := filepath.Dir(fullPath)
			found := false

			// 1. Priority: Check for cover.* in the same directory
			extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
			for _, ext := range extensions {
				coverPath := filepath.Join(dir, "cover"+ext)
				if _, err := os.Stat(coverPath); err == nil {
					fullPath = coverPath
					found = true
					break
				}
			}

			// 2. Priority: Check for any image in the same directory
			if !found {
				entries, err := os.ReadDir(dir)
				if err == nil {
					for _, entry := range entries {
						if !entry.IsDir() && isImageFile(entry.Name()) {
							fullPath = filepath.Join(dir, entry.Name())
							found = true
							break
						}
					}
				}
			}

			// 3. Priority: Check for any image in subdirectories (recursive)
			if !found {
				filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil // Skip errors
					}
					// Found the first image file?
					if !d.IsDir() && isImageFile(d.Name()) {
						fullPath = path
						return fs.SkipAll // Stop walking immediately
					}
					return nil
				})
			}
		}
	}

	// Optimization: Add Cache-Control header
	// Since manga files are static, we can cache them for a long time (e.g., 1 hour or more)
	// This reduces the number of 304 Not Modified checks from the browser
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Use http.ServeFile which handles Content-Type, Range requests, and streaming correctly
	http.ServeFile(w, r, fullPath)
}

// ConvertToCbz converts a directory to a .cbz file and deletes the original directory.
func (s *FileService) ConvertToCbz(dirPath string) error {
	// Clean path
	dirPath = filepath.Clean(dirPath)

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fs.ErrInvalid // Or a custom error "not a directory"
	}

	// Define cbz path
	cbzPath := dirPath + ".cbz"

	// Compress
	err = zipper.CompressDirectory(dirPath, cbzPath)
	if err != nil {
		return err
	}

	// Delete original directory on success
	return os.RemoveAll(dirPath)
}

// DeleteImages deletes specific image files from a directory or a cbz/zip archive.
// It prioritizes: Directory > .cbz > .zip
// relativePath: path relative to the manga directory (e.g. "MangaTitle/Chapter1")
// filenames: list of filenames (basenames) to delete.
func (s *FileService) DeleteImages(relativePath string, filenames []string) error {
	mangaDir, err := s.GetMangaDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(mangaDir, relativePath)

	// 1. Check if it's a directory
	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		// Target is a directory
		for _, fname := range filenames {
			// Basic security check: ensure filename doesn't have path separators
			if strings.Contains(fname, string(os.PathSeparator)) || strings.Contains(fname, "/") {
				continue
			}
			fPath := filepath.Join(fullPath, fname)
			if err := os.Remove(fPath); err != nil {
				// We might want to continue deleting other files even if one fails,
				// or return the error. For now, returning error is safer.
				return err
			}
		}
		return nil
	}

	// 2. Check if .cbz exists
	cbzPath := fullPath + ".cbz"
	if _, err := os.Stat(cbzPath); err == nil {
		return zipper.DeleteFileFromArchive(cbzPath, filenames)
	}

	// 3. Check if .zip exists
	zipPath := fullPath + ".zip"
	if _, err := os.Stat(zipPath); err == nil {
		return zipper.DeleteFileFromArchive(zipPath, filenames)
	}

	return errors.New("target path not found (directory, .cbz, or .zip): " + relativePath)
}

// GetImageList returns a list of image files in a directory or cbz/zip archive.
// It prioritizes: Directory > .cbz > .zip
// relativePath: path relative to the manga directory (e.g. "MangaTitle/Chapter1")
func (s *FileService) GetImageList(relativePath string) ([]string, error) {
	mangaDir, err := s.GetMangaDir()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(mangaDir, relativePath)

	// 1. Check if it's a directory
	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		// Read directory
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, err
		}

		var images []string
		for _, entry := range entries {
			if !entry.IsDir() && isImageFile(entry.Name()) {
				images = append(images, entry.Name())
			}
		}

		// Sort
		sort.Strings(images)
		return images, nil
	}

	// 2. Check if .cbz exists
	cbzPath := fullPath + ".cbz"
	if _, err := os.Stat(cbzPath); err == nil {
		return zipper.ListImages(cbzPath)
	}

	// 3. Check if .zip exists
	zipPath := fullPath + ".zip"
	if _, err := os.Stat(zipPath); err == nil {
		return zipper.ListImages(zipPath)
	}

	// Return empty array if nothing found (as requested)
	return []string{}, nil
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp":
		return true
	}
	return false
}
