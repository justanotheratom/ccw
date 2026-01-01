package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrBranchExists      = errors.New("branch already exists")
	ErrBranchNotMerged   = errors.New("branch not merged")
	ErrBranchNotFound    = errors.New("branch not found")
	ErrRemoteBranchFound = errors.New("remote branch already exists")
)

func BranchExists(repoPath, branch string) (bool, error) {
	_, err := runGit(context.Background(), repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}

	if code, ok := exitCode(err); ok && code == 1 {
		return false, nil
	}

	return false, err
}

func RemoteBranchExists(repoPath, remote, branch string) (bool, error) {
	out, err := runGit(context.Background(), repoPath, "ls-remote", "--heads", remote, branch)
	if err != nil {
		if code, ok := exitCode(err); ok && code == 128 {
			// Treat missing remote as no remote branch found.
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// branchOrRemoteExists checks if a branch exists locally or on origin.
func branchOrRemoteExists(repoPath, branch string) bool {
	if exists, _ := BranchExists(repoPath, branch); exists {
		return true
	}
	// Check remote
	if _, err := runGit(context.Background(), repoPath, "rev-parse", "--verify", "--quiet", "origin/"+branch); err == nil {
		return true
	}
	return false
}

// DetectDefaultBranch auto-detects the default branch (main or master) for a repo.
// Returns an error if both exist or neither exists.
func DetectDefaultBranch(repoPath string) (string, error) {
	mainExists := branchOrRemoteExists(repoPath, "main")
	masterExists := branchOrRemoteExists(repoPath, "master")

	if mainExists && masterExists {
		return "", errors.New("both 'main' and 'master' branches exist; specify --base explicitly")
	}
	if mainExists {
		return "main", nil
	}
	if masterExists {
		return "master", nil
	}
	return "", errors.New("neither 'main' nor 'master' branch found; specify --base explicitly")
}

func resolveBaseRef(repoPath, baseBranch string) (string, error) {
	if baseBranch == "" {
		detected, err := DetectDefaultBranch(repoPath)
		if err != nil {
			return "", err
		}
		baseBranch = detected
	}

	// Prefer remote base branch.
	if _, err := runGit(context.Background(), repoPath, "rev-parse", "--verify", "--quiet", "origin/"+baseBranch); err == nil {
		return "origin/" + baseBranch, nil
	}

	if _, err := runGit(context.Background(), repoPath, "rev-parse", "--verify", "--quiet", baseBranch); err == nil {
		return baseBranch, nil
	}

	return "", fmt.Errorf("base branch %s not found", baseBranch)
}

func CreateBranch(repoPath, branch, baseBranch string, fetch bool) error {
	if fetch {
		if err := Fetch(repoPath, true); err != nil {
			return err
		}
	}

	if exists, err := BranchExists(repoPath, branch); err != nil {
		return err
	} else if exists {
		return ErrBranchExists
	}

	if exists, err := RemoteBranchExists(repoPath, "origin", branch); err != nil {
		return err
	} else if exists {
		return ErrRemoteBranchFound
	}

	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return err
	}

	if _, err := runGit(context.Background(), repoPath, "branch", branch, baseRef); err != nil {
		return err
	}

	return nil
}

func PushBranch(repoPath, branch string) error {
	if _, err := runGit(context.Background(), repoPath, "remote", "get-url", "origin"); err != nil {
		return fmt.Errorf("origin remote not found: %w", err)
	}
	if _, err := runGit(context.Background(), repoPath, "push", "-u", "origin", branch); err != nil {
		return err
	}
	return nil
}

func DeleteRemoteBranch(repoPath, remote, branch string) error {
	if _, err := runGit(context.Background(), repoPath, "push", remote, "--delete", branch); err != nil {
		return err
	}
	return nil
}

func DeleteBranch(repoPath, branch string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D", branch)
	} else {
		args = append(args, "-d", branch)
	}

	if _, err := runGit(context.Background(), repoPath, args...); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrBranchNotFound
		}
		if !force && strings.Contains(err.Error(), "not fully merged") {
			return ErrBranchNotMerged
		}
		return err
	}

	return nil
}

func IsMerged(repoPath, branch, baseBranch string, fetch bool) (bool, error) {
	if fetch {
		if err := Fetch(repoPath, true); err != nil {
			return false, err
		}
	}

	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return false, err
	}

	// Fast path: check if branch is an ancestor of base (regular merge)
	_, err = runGit(context.Background(), repoPath, "merge-base", "--is-ancestor", branch, baseRef)
	if err == nil {
		return true, nil
	}

	if code, ok := exitCode(err); ok && code == 1 {
		// Not an ancestor - could be a squash merge. Check if the diff is empty,
		// which means all changes from the branch are already in base.
		_, diffErr := runGit(context.Background(), repoPath, "diff", "--quiet", branch, baseRef)
		if diffErr == nil {
			// No diff means effectively merged (squash merge case)
			return true, nil
		}
		return false, nil
	}

	return false, err
}

func HasUnpushedCommits(repoPath, branch string) (bool, error) {
	if _, err := runGit(context.Background(), repoPath, "rev-parse", "--verify", "--quiet", "origin/"+branch); err != nil {
		// If remote branch is missing, treat all commits as unpushed.
		return true, nil
	}

	out, err := runGit(context.Background(), repoPath, "rev-list", "--left-only", "--count", branch+"...origin/"+branch)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(out) != "0", nil
}

// RemoteBranchHasUnmergedCommits checks if the remote branch has commits not in the base branch.
// This is useful before deleting a remote branch to ensure no work is lost.
// Returns false if the remote branch doesn't exist.
func RemoteBranchHasUnmergedCommits(repoPath, branch, baseBranch string) (bool, error) {
	remoteBranch := "origin/" + branch

	// Check if remote branch exists
	if _, err := runGit(context.Background(), repoPath, "rev-parse", "--verify", "--quiet", remoteBranch); err != nil {
		// Remote branch doesn't exist, nothing to check
		return false, nil
	}

	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return false, err
	}

	// Count commits in origin/<branch> that are not in baseRef
	out, err := runGit(context.Background(), repoPath, "rev-list", "--count", baseRef+".."+remoteBranch)
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(out) == "0" {
		// No commits difference by ancestry
		return false, nil
	}

	// Has commits not in base by ancestry, but could be a squash merge.
	// Check if the diff is empty (all changes incorporated).
	_, diffErr := runGit(context.Background(), repoPath, "diff", "--quiet", remoteBranch, baseRef)
	if diffErr == nil {
		// No diff means effectively merged (squash merge case)
		return false, nil
	}

	return true, nil
}

// GetDiffFiles returns the list of files that differ between a branch and base.
// Returns nil if there are no differences.
func GetDiffFiles(repoPath, branch, baseBranch string) ([]string, error) {
	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return nil, err
	}

	out, err := runGit(context.Background(), repoPath, "diff", "--name-only", branch, baseRef)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}

	return strings.Split(trimmed, "\n"), nil
}

func TouchBranch(repoPath, branch string) error {
	_, err := runGit(context.Background(), repoPath, "update-ref", "--no-deref", "--create-reflog", "refs/heads/"+branch, "HEAD")
	return err
}

func Now() time.Time {
	return time.Now().UTC()
}
