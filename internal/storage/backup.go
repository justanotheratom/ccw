package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupFile copies the file at path to a timestamped backup in the same
// directory. Returns the backup path or an empty string if the source file does
// not exist.
func BackupFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("cannot back up directory: %s", path)
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	backupName := fmt.Sprintf("%s.bak-%s", base, timestamp)
	backupPath := filepath.Join(dir, backupName)

	if err := copyFile(path, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}
