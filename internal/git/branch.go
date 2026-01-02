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

// MergeChecker is a function that checks if a branch was merged via PR.
// Returns (merged, found, error) where found=false means no PR exists for this branch.
type MergeChecker func(ctx context.Context, branch string) (merged bool, found bool, err error)

// IsMergedWithPR checks if a branch is merged, optionally using PR detection first.
// If prChecker is provided and finds a PR, its result is used. Otherwise falls back to git heuristics.
func IsMergedWithPR(ctx context.Context, repoPath, branch, baseBranch string, fetch bool, prChecker MergeChecker) (bool, error) {
	if fetch {
		if err := Fetch(repoPath, true); err != nil {
			return false, err
		}
	}

	// Try PR-based detection first if available
	if prChecker != nil {
		merged, found, err := prChecker(ctx, branch)
		if err == nil && found {
			return merged, nil
		}
		// Fall through to git-based detection on error or not found
	}

	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return false, err
	}

	// Fast path: check if branch is an ancestor of base (regular merge)
	_, err = runGit(ctx, repoPath, "merge-base", "--is-ancestor", branch, baseRef)
	if err == nil {
		return true, nil
	}

	if code, ok := exitCode(err); ok && code == 1 {
		// Not an ancestor - could be a squash merge. Check if the branch's changes
		// are in base by comparing only the files that the branch modified.
		mergeBase, mbErr := runGit(ctx, repoPath, "merge-base", branch, baseRef)
		if mbErr != nil {
			return false, nil
		}
		mergeBase = strings.TrimSpace(mergeBase)

		// Get files that branch changed vs merge-base
		changedFiles, cfErr := runGit(ctx, repoPath, "diff", "--name-only", mergeBase, branch)
		if cfErr != nil || strings.TrimSpace(changedFiles) == "" {
			// No changes on branch, consider it merged
			return true, nil
		}

		// Check if those specific files are the same in branch and base
		files := strings.Fields(changedFiles)
		args := append([]string{"diff", "--quiet", branch, baseRef, "--"}, files...)
		_, diffErr := runGit(ctx, repoPath, args...)
		if diffErr == nil {
			// Branch's changed files match base - squash merged
			return true, nil
		}
		return false, nil
	}

	return false, err
}

// IsMerged checks if a branch is merged into the base branch using git heuristics only.
func IsMerged(repoPath, branch, baseBranch string, fetch bool) (bool, error) {
	return IsMergedWithPR(context.Background(), repoPath, branch, baseBranch, fetch, nil)
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

// RemoteBranchHasUnmergedCommitsWithPR checks if the remote branch has commits not in the base branch.
// Uses PR checker first if provided, falls back to git heuristics.
// Returns false if the remote branch doesn't exist.
func RemoteBranchHasUnmergedCommitsWithPR(ctx context.Context, repoPath, branch, baseBranch string, prChecker MergeChecker) (bool, error) {
	remoteBranch := "origin/" + branch

	// Check if remote branch exists
	if _, err := runGit(ctx, repoPath, "rev-parse", "--verify", "--quiet", remoteBranch); err != nil {
		// Remote branch doesn't exist, nothing to check
		return false, nil
	}

	// Try PR-based detection first if available
	if prChecker != nil {
		merged, found, err := prChecker(ctx, branch)
		if err == nil && found {
			return !merged, nil // Has unmerged = !merged
		}
		// Fall through to git-based detection on error or not found
	}

	baseRef, err := resolveBaseRef(repoPath, baseBranch)
	if err != nil {
		return false, err
	}

	// Count commits in origin/<branch> that are not in baseRef
	out, err := runGit(ctx, repoPath, "rev-list", "--count", baseRef+".."+remoteBranch)
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(out) == "0" {
		// No commits difference by ancestry
		return false, nil
	}

	// Has commits not in base by ancestry, but could be a squash merge.
	// Check if the remote branch's changes are in base.
	mergeBase, mbErr := runGit(ctx, repoPath, "merge-base", remoteBranch, baseRef)
	if mbErr != nil {
		return true, nil
	}
	mergeBase = strings.TrimSpace(mergeBase)

	// Get files that remote branch changed vs merge-base
	changedFiles, cfErr := runGit(ctx, repoPath, "diff", "--name-only", mergeBase, remoteBranch)
	if cfErr != nil || strings.TrimSpace(changedFiles) == "" {
		// No changes on remote branch
		return false, nil
	}

	// Check if those specific files are the same in remote branch and base
	files := strings.Fields(changedFiles)
	args := append([]string{"diff", "--quiet", remoteBranch, baseRef, "--"}, files...)
	_, diffErr := runGit(ctx, repoPath, args...)
	if diffErr == nil {
		// Remote branch's changed files match base - squash merged
		return false, nil
	}

	return true, nil
}

// RemoteBranchHasUnmergedCommits checks if the remote branch has commits not in the base branch.
// This is useful before deleting a remote branch to ensure no work is lost.
// Returns false if the remote branch doesn't exist.
func RemoteBranchHasUnmergedCommits(repoPath, branch, baseBranch string) (bool, error) {
	return RemoteBranchHasUnmergedCommitsWithPR(context.Background(), repoPath, branch, baseBranch, nil)
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
