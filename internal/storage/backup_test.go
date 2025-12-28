package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupFileCreatesCopy(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "file.txt")
	content := []byte("hello")
	if err := os.WriteFile(sourcePath, content, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	backupPath, err := BackupFile(sourcePath)
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}

	if backupPath == "" {
		t.Fatalf("expected backup path")
	}

	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("backup file missing: %v", err)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}

	if string(data) != string(content) {
		t.Fatalf("expected backup to match content")
	}

	if !strings.Contains(backupPath, ".bak-") {
		t.Fatalf("expected timestamp suffix, got %s", backupPath)
	}
}

func TestBackupFileMissingSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")
	backup, err := BackupFile(path)
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}
	if backup != "" {
		t.Fatalf("expected empty backup path for missing source, got %s", backup)
	}
}
