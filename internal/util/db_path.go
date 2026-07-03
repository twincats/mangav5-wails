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

	dbFileName := "mangav5.dev.db"
	if IsProductionBuild() {
		dbFileName = "mangav5.db"
	}
	dbPath := filepath.Join(appDir, dbFileName)
	log.Println("Database path:", dbPath)

	if IsProductionBuild() {
		localDB := "manga.db"
		if _, err := os.Stat(localDB); err == nil {
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				log.Println("Found local database, migrating to AppData...")
				if err := copyFile(localDB, dbPath); err != nil {
					log.Println("Failed to write to new database location:", err)
				} else {
					log.Println("Database migrated successfully.")
					os.Rename(localDB, localDB+".bak")
				}
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
