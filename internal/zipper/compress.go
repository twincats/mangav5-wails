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

// ReadFileFromArchive reads a specific file from a zip/cbz archive.
// It searches for the file by its path inside the archive.
func ReadFileFromArchive(archivePath string, filePathInZip string) ([]byte, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		// Use exact match or check for path normalization if needed.
		// Usually zip paths use forward slashes.
		if f.Name == filePathInZip {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, errors.New("file not found in archive: " + filePathInZip)
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
