package services

import (
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
// targetPath: path to the directory or the archive file.
// filenames: list of filenames (basenames) to delete.
func (s *FileService) DeleteImages(targetPath string, filenames []string) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Target is a directory
		for _, fname := range filenames {
			// Basic security check: ensure filename doesn't have path separators
			if strings.Contains(fname, string(os.PathSeparator)) || strings.Contains(fname, "/") {
				continue
			}
			fullPath := filepath.Join(targetPath, fname)
			if err := os.Remove(fullPath); err != nil {
				// We might want to continue deleting other files even if one fails,
				// or return the error. For now, returning error is safer.
				return err
			}
		}
		return nil
	}

	// Target is a file (assume archive)
	// We should probably check extension, but zipper.DeleteFileFromArchive handles zip/cbz.
	return zipper.DeleteFileFromArchive(targetPath, filenames)
}

// GetImageList returns a list of image files in a directory or cbz/zip archive.
// Returns filenames sorted alphabetically.
func (s *FileService) GetImageList(targetPath string) ([]string, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// Read directory
		entries, err := os.ReadDir(targetPath)
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

	// Target is a file (assume archive)
	return zipper.ListImages(targetPath)
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp":
		return true
	}
	return false
}
