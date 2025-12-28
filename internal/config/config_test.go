package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Version != CurrentVersion {
		t.Fatalf("expected version %d, got %d", CurrentVersion, cfg.Version)
	}

	if cfg.ReposDir != defaultReposDir {
		t.Fatalf("expected repos dir %q, got %q", defaultReposDir, cfg.ReposDir)
	}

	if _, err := os.Stat(store.Path()); err != nil {
		t.Fatalf("expected config file to be written: %v", err)
	}
}

func TestConfigLoadExisting(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	want := Config{
		Version:           CurrentVersion,
		ReposDir:          "~/projects",
		DefaultBase:       "develop",
		ITermCCMode:       false,
		ClaudeRenameDelay: 3,
		Layout: Layout{
			Left:  "claude",
			Right: "custom",
		},
	}

	data, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := os.WriteFile(store.Path(), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.DefaultBase != want.DefaultBase || got.ReposDir != want.ReposDir || got.ITermCCMode != want.ITermCCMode {
		t.Fatalf("loaded config mismatch: %+v", got)
	}
}

func TestConfigSave(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg := Default()
	cfg.ReposDir = "~/workspace"

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ReposDir != cfg.ReposDir {
		t.Fatalf("expected repos dir %q, got %q", cfg.ReposDir, decoded.ReposDir)
	}
}

func TestExpandTilde(t *testing.T) {
	path, err := ExpandPath("~/something")
	if err != nil {
		t.Fatalf("ExpandPath: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got %q", path)
	}

	if path == "~/something" {
		t.Fatalf("expected tilde expansion, got %q", path)
	}
}

func TestExpandedReposDir(t *testing.T) {
	cfg := Default()
	cfg.ReposDir = "~/github"
	expanded, err := cfg.ExpandedReposDir()
	if err != nil {
		t.Fatalf("ExpandedReposDir: %v", err)
	}
	if expanded == cfg.ReposDir {
		t.Fatalf("expected expanded path, got %q", expanded)
	}
}

func TestConfigSaveCreatesBackup(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg := Default()
	if err := store.Save(cfg); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	cfg.DefaultBase = "develop"
	if err := store.Save(cfg); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	var backupFound bool
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "config.json.bak-") {
			backupFound = true
			break
		}
	}

	if !backupFound {
		t.Fatalf("expected backup file to be created")
	}
}
