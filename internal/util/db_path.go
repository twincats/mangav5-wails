package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// GetAndMigrateDatabasePath determines the database path in AppData
// and migrates the local database if it exists and the target does not.
func GetAndMigrateDatabasePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "mangav5")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create application directory: %w", err)
	}

	dbPath := filepath.Join(appDir, "mangav5.db")
	log.Println("Database path:", dbPath)

	// Migration: If manga.db exists in current directory but not in AppData, copy it.
	localDB := "manga.db"
	if _, err := os.Stat(localDB); err == nil {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			log.Println("Found local database, migrating to AppData...")
			if err := copyFile(localDB, dbPath); err != nil {
				log.Println("Failed to write to new database location:", err)
			} else {
				log.Println("Database migrated successfully.")
				// Optional: Rename old db to .bak to avoid confusion
				os.Rename(localDB, localDB+".bak")
			}
		}
	}

	return dbPath, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
