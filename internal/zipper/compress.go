package zipper

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauspost/compress/zip"
)

func CompressDirectory(sourceDir string, destZipFile string) error {
	sourceDir = filepath.Clean(sourceDir)

	info, err := os.Stat(sourceDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("sourceDir is not a directory")
	}

	// Create a new zip file
	out, err := os.Create(destZipFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Use bufio for better I/O performance (32KB buffer)
	bufferedWriter := bufio.NewWriterSize(out, 32*1024)
	defer bufferedWriter.Flush()

	// Create a new zip writer
	zipWriter := zip.NewWriter(bufferedWriter)
	defer zipWriter.Close()

	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		name := filepath.ToSlash(relPath)

		fi, err := d.Info()
		if err != nil {
			return err
		}

		hdr, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		hdr.Name = name

		// Optimize: Store already compressed files, Deflate others
		if isCompressed(name) {
			hdr.Method = zip.Store
		} else {
			hdr.Method = zip.Deflate
		}

		// Note: hdr.Modified is already set by zip.FileInfoHeader(fi)
		// We do not zero it out anymore to preserve timestamps.

		zipFile, err := zipWriter.CreateHeader(hdr)
		if err != nil {
			return err
		}
		// Copy the file content to the zip file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	})

}

// ListImages returns a list of image filenames inside a zip/cbz archive.
// It supports .zip and .cbz files.
func ListImages(archivePath string) ([]string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var images []string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if isImage(f.Name) {
			images = append(images, f.Name)
		}
	}

	// Sort images by name to ensure consistent order
	sort.Strings(images)

	return images, nil
}

// OpenFileFromArchive opens a file stream from a zip/cbz archive.
// Caller is responsible for closing the returned ReadCloser.
// Uses an LRU cache to avoid reopening the zip file for every request.
func OpenFileFromArchive(archivePath string, filePathInZip string) (io.ReadCloser, error) {
	// Use the global cache to get an open reader
	r, err := GetZipCache().GetOrOpen(archivePath)
	if err != nil {
		return nil, err
	}
	// Note: We DO NOT defer r.Close() here because the cache manages the lifecycle.

	for _, f := range r.File {
		// Use exact match or check for path normalization if needed.
		if f.Name == filePathInZip {
			return f.Open()
		}
	}

	return nil, errors.New("file not found in archive: " + filePathInZip)
}

// DeleteFileFromArchive deletes specific files from a compressed chapter archive (CBZ/ZIP).
// It creates a new temporary zip file, copies all files except the ones to be deleted,
// and then replaces the original file.
func DeleteFileFromArchive(archivePath string, filesToDelete []string) error {
	// 1. Remove from Cache (Critical for Windows)
	GetZipCache().Remove(archivePath)

	// 2. Open source zip
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 3. Create temp file
	tempFile, err := os.CreateTemp(filepath.Dir(archivePath), "temp_*.zip")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up if something fails

	// 4. Create zip writer
	w := zip.NewWriter(tempFile)

	// Prepare deletion set for O(1) lookup
	toDelete := make(map[string]bool)
	for _, f := range filesToDelete {
		toDelete[f] = true
	}

	// 5. Iterate and copy
	for _, f := range r.File {
		if toDelete[f.Name] {
			continue // Skip deleted files
		}

		// Copy file to new zip
		rc, err := f.Open()
		if err != nil {
			w.Close()
			tempFile.Close()
			return err
		}

		// Create header (preserve metadata)
		header := f.FileHeader
		writer, err := w.CreateHeader(&header)
		if err != nil {
			rc.Close()
			w.Close()
			tempFile.Close()
			return err
		}

		_, err = io.Copy(writer, rc)
		rc.Close()
		if err != nil {
			w.Close()
			tempFile.Close()
			return err
		}
	}

	// 6. Close everything
	if err := w.Close(); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	// Close reader explicitly before renaming (for Windows)
	r.Close()

	// 7. Replace original file
	return os.Rename(tempPath, archivePath)
}

// ExtractFileFromArchive extracts a single file from a compressed chapter archive (CBZ/ZIP)
// to the specified destination path.
func ExtractFileFromArchive(archivePath string, filePathInZip string, destPath string) error {
	// Use Cache
	rc, err := OpenFileFromArchive(archivePath, filePathInZip)
	if err != nil {
		return err
	}
	defer rc.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy content
	_, err = io.Copy(out, rc)
	return err
}

// ExtractAllFromArchive extracts all files from a compressed chapter archive (CBZ/ZIP)
// to the specified destination directory.
func ExtractAllFromArchive(archivePath string, destDir string) error {
	// We do NOT use the cache here because we need to iterate all files,
	// and OpenReader gives us easy access to the File slice.
	// Also, this is a heavy operation, so opening a fresh reader is fine.
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Zip Slip protection
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return errors.New("illegal file path: " + fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create directory tree
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Create file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// isCompressed checks if the file extension suggests the file is already compressed.
func isCompressed(filename string) bool {
	if isImage(filename) {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".zip", ".rar", ".7z", ".gz", ".mp4", ".mkv", ".avi", ".mov":
		return true
	}
	return false
}

// isImage checks if the file is an image based on extension.
func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".avif":
		return true
	}
	return false
}
