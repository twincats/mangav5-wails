package services

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// FileService handles QR code generation
type FileService struct {
	dbService *DatabaseService
}

var MANGA_DIR = ""

// NewQRService creates a new QR service
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

	// Fallback logic for cover.webp
	// If cover.webp is requested but does not exist, try to serve the first image in the directory or subdirectories
	if strings.EqualFold(filepath.Base(fullPath), "cover.webp") {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			dir := filepath.Dir(fullPath)

			// WalkDir traverses the directory tree rooted at dir
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

	// Optimization: Add Cache-Control header
	// Since manga files are static, we can cache them for a long time (e.g., 1 hour or more)
	// This reduces the number of 304 Not Modified checks from the browser
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Use http.ServeFile which handles Content-Type, Range requests, and streaming correctly
	http.ServeFile(w, r, fullPath)
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp":
		return true
	}
	return false
}
