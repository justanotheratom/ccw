package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ccw/ccw/internal/storage"
	"github.com/gofrs/flock"
)

const (
	registryFileName = "workspaces.json"
	lockFileName     = "workspaces.json.lock"
	CurrentVersion   = 1
	lockRetry        = 50 * time.Millisecond
)

var ErrUnsupportedVersion = errors.New("unsupported registry version")

type Workspace struct {
	Repo           string    `json:"repo"`
	RepoPath       string    `json:"repo_path"`
	Branch         string    `json:"branch"`
	BaseBranch     string    `json:"base_branch"`
	WorktreePath   string    `json:"worktree_path"`
	ClaudeSession  string    `json:"claude_session"`
	TmuxSession    string    `json:"tmux_session"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

type Registry struct {
	Version    int                  `json:"version"`
	Workspaces map[string]Workspace `json:"workspaces"`
}

type Store struct {
	root        string
	lockTimeout time.Duration
}

func NewStore(root string) (*Store, error) {
	return newStore(root, 5*time.Second)
}

func newStore(root string, timeout time.Duration) (*Store, error) {
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home directory: %w", err)
		}
		root = filepath.Join(home, ".ccw")
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve registry path: %w", err)
	}

	return &Store{
		root:        abs,
		lockTimeout: timeout,
	}, nil
}

func defaultRegistry() Registry {
	return Registry{
		Version:    CurrentVersion,
		Workspaces: map[string]Workspace{},
	}
}

func (s *Store) registryPath() string {
	return filepath.Join(s.root, registryFileName)
}

func (s *Store) lockPath() string {
	return filepath.Join(s.root, lockFileName)
}

// Read loads the registry without acquiring a write lock. Callers should
// prefer Update when modifying the registry to ensure exclusive access.
func (s *Store) Read(ctx context.Context) (Registry, error) {
	ctx, cancel := context.WithTimeout(ctx, s.lockTimeout)
	defer cancel()

	lock := flock.New(s.lockPath())
	locked, err := lock.TryRLockContext(ctx, lockRetry)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Registry{}, fmt.Errorf("acquire read lock: timed out after %s", s.lockTimeout)
		}
		return Registry{}, fmt.Errorf("acquire read lock: %w", err)
	}
	if !locked {
		return Registry{}, fmt.Errorf("acquire read lock: timed out after %s", s.lockTimeout)
	}
	defer lock.Unlock()

	return s.load()
}

// Update acquires an exclusive lock, loads the registry, allows the caller to
// mutate it, then saves the result atomically.
func (s *Store) Update(ctx context.Context, fn func(*Registry) error) error {
	ctx, cancel := context.WithTimeout(ctx, s.lockTimeout)
	defer cancel()

	lock := flock.New(s.lockPath())
	locked, err := lock.TryLockContext(ctx, lockRetry)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("acquire write lock: timed out after %s", s.lockTimeout)
		}
		return fmt.Errorf("acquire write lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("acquire write lock: timed out after %s", s.lockTimeout)
	}
	defer lock.Unlock()

	reg, err := s.load()
	if err != nil {
		return err
	}

	if err := fn(&reg); err != nil {
		return err
	}

	return s.save(reg)
}

func (s *Store) load() (Registry, error) {
	path := s.registryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultRegistry(), nil
		}
		return Registry{}, fmt.Errorf("read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		backupReg, _, backupErr := s.loadLatestBackup()
		if backupErr == nil {
			// Recover silently from latest backup.
			return backupReg, nil
		}
		return Registry{}, fmt.Errorf("parse registry and no usable backup: %w", err)
	}

	if reg.Version != CurrentVersion {
		return Registry{}, ErrUnsupportedVersion
	}

	if reg.Workspaces == nil {
		reg.Workspaces = map[string]Workspace{}
	}

	return reg, nil
}

func (s *Store) save(reg Registry) error {
	if reg.Version == 0 {
		reg.Version = CurrentVersion
	}

	if reg.Workspaces == nil {
		reg.Workspaces = map[string]Workspace{}
	}

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	if _, err := storage.BackupFile(s.registryPath()); err != nil {
		return fmt.Errorf("backup existing registry: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}

	tmpPath := s.registryPath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp registry: %w", err)
	}

	if err := os.Rename(tmpPath, s.registryPath()); err != nil {
		return fmt.Errorf("atomically write registry: %w", err)
	}

	return nil
}

func (s *Store) loadLatestBackup() (Registry, string, error) {
	dirEntries, err := os.ReadDir(s.root)
	if err != nil {
		return Registry{}, "", err
	}

	var latestPath string
	var latestInfo os.FileInfo
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), registryFileName+".bak-") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if latestInfo == nil || info.ModTime().After(latestInfo.ModTime()) {
				latestInfo = info
				latestPath = filepath.Join(s.root, entry.Name())
			}
		}
	}

	if latestPath == "" {
		return Registry{}, "", fmt.Errorf("no backup found")
	}

	data, err := os.ReadFile(latestPath)
	if err != nil {
		return Registry{}, "", err
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return Registry{}, "", err
	}

	return reg, latestPath, nil
}

// Add inserts a workspace into the registry under the given ID. Returns an
// error if the ID already exists.
func (r *Registry) Add(id string, ws Workspace) error {
	if r.Workspaces == nil {
		r.Workspaces = map[string]Workspace{}
	}

	if _, exists := r.Workspaces[id]; exists {
		return fmt.Errorf("workspace %s already exists", id)
	}

	r.Workspaces[id] = ws
	return nil
}

func (r *Registry) Remove(id string) {
	delete(r.Workspaces, id)
}

func (r *Registry) Get(id string) (Workspace, bool) {
	ws, ok := r.Workspaces[id]
	return ws, ok
}

// FindByPartialName returns workspace IDs matching the provided prefix in a
// case-insensitive manner.
func (r *Registry) FindByPartialName(partial string) []string {
	matches := []string{}
	partial = strings.ToLower(partial)
	for id := range r.Workspaces {
		if strings.Contains(strings.ToLower(id), partial) {
			matches = append(matches, id)
		}
	}
	return matches
}
