package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	if _, err := runGit(context.Background(), dir, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}

	if _, err := runGit(context.Background(), dir, "checkout", "-b", "main"); err != nil {
		t.Fatalf("git checkout: %v", err)
	}

	if _, err := runGit(context.Background(), dir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name: %v", err)
	}

	if _, err := runGit(context.Background(), dir, "commit", "--allow-empty", "-m", "initial"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	return dir
}

func TestValidateRepoExists(t *testing.T) {
	repo := initRepo(t)
	path, err := ValidateRepo(repo)
	if err != nil {
		t.Fatalf("ValidateRepo: %v", err)
	}
	if path != repo {
		t.Fatalf("expected %s, got %s", repo, path)
	}
}

func TestValidateRepoNotFound(t *testing.T) {
	if _, err := ValidateRepo("/tmp/does-not-exist"); err == nil {
		t.Fatalf("expected error for missing repo")
	}
}

func TestCreateBranchSuccess(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	exists, err := BranchExists(repo, "feature/test")
	if err != nil || !exists {
		t.Fatalf("expected branch to exist after creation")
	}
}

func TestCreateBranchAlreadyExists(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch first: %v", err)
	}

	err := CreateBranch(repo, "feature/test", "main", false)
	if err == nil || err != ErrBranchExists {
		t.Fatalf("expected ErrBranchExists, got %v", err)
	}
}

func TestDeleteBranchSuccess(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	if err := DeleteBranch(repo, "feature/test", false); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}
}

func TestDeleteBranchNotMergedError(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// add unmerged commit
	if _, err := runGit(context.Background(), repo, "checkout", "feature/test"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("change"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "-m", "feature work"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "checkout", "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	err := DeleteBranch(repo, "feature/test", false)
	if err == nil || err != ErrBranchNotMerged {
		t.Fatalf("expected ErrBranchNotMerged, got %v", err)
	}
}

func TestDeleteBranchForce(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	if err := DeleteBranch(repo, "feature/test", true); err != nil {
		t.Fatalf("DeleteBranch force: %v", err)
	}
}

func TestIsBranchMergedTrue(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	merged, err := IsMerged(repo, "feature/test", "main", false)
	if err != nil {
		t.Fatalf("IsMerged: %v", err)
	}
	if !merged {
		t.Fatalf("expected branch to be considered merged (no new commits)")
	}
}

func TestIsBranchMergedFalse(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	if _, err := runGit(context.Background(), repo, "checkout", "feature/test"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "--allow-empty", "-m", "feature change"); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "checkout", "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	merged, err := IsMerged(repo, "feature/test", "main", false)
	if err != nil {
		t.Fatalf("IsMerged: %v", err)
	}
	if merged {
		t.Fatalf("expected branch to be unmerged")
	}
}

func TestCreateWorktreeSuccess(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	worktreePath := filepath.Join(t.TempDir(), "worktree")
	if err := CreateWorktree(repo, worktreePath, "feature/test"); err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	exists, err := WorktreeExists(repo, worktreePath)
	if err != nil {
		t.Fatalf("WorktreeExists: %v", err)
	}
	if !exists {
		t.Fatalf("expected worktree to exist")
	}
}

func TestRemoveWorktreeSuccess(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	worktreePath := filepath.Join(t.TempDir(), "worktree")
	if err := CreateWorktree(repo, worktreePath, "feature/test"); err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	if err := RemoveWorktree(repo, worktreePath, true); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}

	if _, err := os.Stat(worktreePath); err == nil {
		t.Fatalf("expected worktree directory to be removed")
	}
}

func TestRemoveWorktreeMissingIsOk(t *testing.T) {
	repo := initRepo(t)
	err := RemoveWorktree(repo, "/tmp/ccw-missing-worktree", true)
	if err != nil {
		t.Fatalf("expected no error for missing worktree, got %v", err)
	}
}
