package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ccw/ccw/internal/config"
)

var ErrRepoNotFound = fmt.Errorf("repository not found")

func ValidateRepo(repoPath string) (string, error) {
	expanded, err := config.ExpandPath(repoPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrRepoNotFound
		}
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", expanded)
	}

	if _, err := runGit(context.Background(), expanded, "rev-parse", "--is-inside-work-tree"); err != nil {
		return "", fmt.Errorf("not a git repository: %s", expanded)
	}

	return expanded, nil
}

func Fetch(repoPath string, prune bool) error {
	args := []string{"fetch"}
	if prune {
		args = append(args, "--prune")
	}

	_, err := runGit(context.Background(), repoPath, args...)
	return err
}

func DefaultWorktreePath(rootDir, safeName string) (string, error) {
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		rootDir = filepath.Join(home, ".ccw", "worktrees")
	}

	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return "", fmt.Errorf("create worktree root: %w", err)
	}

	return filepath.Join(rootDir, safeName), nil
}
