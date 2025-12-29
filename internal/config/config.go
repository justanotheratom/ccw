package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ccw/ccw/internal/storage"
)

const (
	CurrentVersion  = 1
	dirName         = ".ccw"
	configFileName  = "config.json"
	defaultReposDir = "~/github"
)

var ErrUnsupportedVersion = errors.New("unsupported config version")

type Layout struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

type Config struct {
	Version                    int    `json:"version"`
	ReposDir                   string `json:"repos_dir"`
	ITermCCMode                bool   `json:"iterm_cc_mode"`
	ClaudeRenameDelay          int    `json:"claude_rename_delay"`
	Layout                     Layout `json:"layout"`
	Onboarded                  bool   `json:"onboarded"`
	ClaudeDangerouslySkipPerms bool   `json:"claude_dangerously_skip_permissions"`
}

type Store struct {
	root string
}

func NewStore(root string) (*Store, error) {
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home directory: %w", err)
		}
		root = filepath.Join(home, dirName)
	}

	expanded, err := ExpandPath(root)
	if err != nil {
		return nil, err
	}

	return &Store{root: expanded}, nil
}

func Default() Config {
	return Config{
		Version:                    CurrentVersion,
		ReposDir:                   defaultReposDir,
		ITermCCMode:                true,
		ClaudeRenameDelay:          5,
		Layout:                     Layout{Left: "claude", Right: "lazygit"},
		Onboarded:                  false,
		ClaudeDangerouslySkipPerms: false,
	}
}

func (s *Store) Path() string {
	return filepath.Join(s.root, configFileName)
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) Load() (Config, error) {
	path := s.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if err := s.Save(cfg); err != nil {
				return Config{}, err
			}
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Version != CurrentVersion {
		return Config{}, ErrUnsupportedVersion
	}

	return cfg, nil
}

func (s *Store) Save(cfg Config) error {
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if _, err := storage.BackupFile(s.Path()); err != nil {
		return fmt.Errorf("backup existing config: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	tmpPath := s.Path() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}

	if err := os.Rename(tmpPath, s.Path()); err != nil {
		return fmt.Errorf("atomically write config: %w", err)
	}

	return nil
}

func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	return abs, nil
}

func (c Config) ExpandedReposDir() (string, error) {
	return ExpandPath(c.ReposDir)
}
