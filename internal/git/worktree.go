package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WorktreeExists(repoPath, path string) (bool, error) {
	out, err := runGit(context.Background(), repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return false, err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return false, err
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	path = filepath.Clean(path)

	entries := strings.Split(out, "\n")
	for _, line := range entries {
		if strings.HasPrefix(line, "worktree ") {
			wtPath := filepath.Clean(strings.TrimPrefix(line, "worktree "))
			if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
				wtPath = resolved
			}
			if wtPath == path {
				return true, nil
			}
		}
	}
	return false, nil
}

func CreateWorktree(repoPath, path, branch string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create worktree parent dir: %w", err)
	}

	args := []string{"worktree", "add", path, branch}
	_, err := runGit(context.Background(), repoPath, args...)
	return err
}

func RemoveWorktree(repoPath, path string, force bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	_, err := runGit(context.Background(), repoPath, args...)
	if err != nil && strings.Contains(err.Error(), "is not a working tree") {
		return nil
	}
	return err
}
