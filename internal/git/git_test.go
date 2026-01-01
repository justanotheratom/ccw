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
	// Add an actual file change (not empty commit) to test unmerged detection
	if err := os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "add", "feature.txt"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "-m", "feature change"); err != nil {
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

func TestIsBranchMergedSquash(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// Add a change on the feature branch
	if _, err := runGit(context.Background(), repo, "checkout", "feature/test"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "add", "feature.txt"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "-m", "feature change"); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Simulate squash merge: apply same changes to main with a different commit
	if _, err := runGit(context.Background(), repo, "checkout", "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "add", "feature.txt"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "-m", "squash: feature change"); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// The feature branch is not an ancestor of main, but the diff is empty
	// so it should be detected as effectively merged
	merged, err := IsMerged(repo, "feature/test", "main", false)
	if err != nil {
		t.Fatalf("IsMerged: %v", err)
	}
	if !merged {
		t.Fatalf("expected squash-merged branch to be considered merged")
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

func TestDetectDefaultBranchMain(t *testing.T) {
	repo := initRepo(t) // initRepo creates a repo with main branch
	branch, err := DetectDefaultBranch(repo)
	if err != nil {
		t.Fatalf("DetectDefaultBranch: %v", err)
	}
	if branch != "main" {
		t.Fatalf("expected main, got %s", branch)
	}
}

func TestDetectDefaultBranchMaster(t *testing.T) {
	dir := t.TempDir()

	if _, err := runGit(context.Background(), dir, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "checkout", "-b", "master"); err != nil {
		t.Fatalf("git checkout: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "commit", "--allow-empty", "-m", "initial"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	branch, err := DetectDefaultBranch(dir)
	if err != nil {
		t.Fatalf("DetectDefaultBranch: %v", err)
	}
	if branch != "master" {
		t.Fatalf("expected master, got %s", branch)
	}
}

func TestDetectDefaultBranchNeitherError(t *testing.T) {
	dir := t.TempDir()

	if _, err := runGit(context.Background(), dir, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "checkout", "-b", "develop"); err != nil {
		t.Fatalf("git checkout: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	if _, err := runGit(context.Background(), dir, "commit", "--allow-empty", "-m", "initial"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	_, err := DetectDefaultBranch(dir)
	if err == nil {
		t.Fatal("expected error when neither main nor master exists")
	}
}

// initRepoWithRemote creates a local repo with a bare "origin" remote for testing.
func initRepoWithRemote(t *testing.T) (localRepo, bareRemote string) {
	t.Helper()

	// Create bare remote
	bareRemote = t.TempDir()
	if _, err := runGit(context.Background(), bareRemote, "init", "--bare"); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}

	// Create local repo
	localRepo = t.TempDir()
	if _, err := runGit(context.Background(), localRepo, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "checkout", "-b", "main"); err != nil {
		t.Fatalf("git checkout: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "commit", "--allow-empty", "-m", "initial"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Add bare repo as origin
	if _, err := runGit(context.Background(), localRepo, "remote", "add", "origin", bareRemote); err != nil {
		t.Fatalf("git remote add: %v", err)
	}

	// Push main to origin
	if _, err := runGit(context.Background(), localRepo, "push", "-u", "origin", "main"); err != nil {
		t.Fatalf("git push: %v", err)
	}

	return localRepo, bareRemote
}

func TestRemoteBranchHasUnmergedCommits_NoRemote(t *testing.T) {
	repo := initRepo(t)

	// No remote exists, should return false
	hasUnmerged, err := RemoteBranchHasUnmergedCommits(repo, "feature/test", "main")
	if err != nil {
		t.Fatalf("RemoteBranchHasUnmergedCommits: %v", err)
	}
	if hasUnmerged {
		t.Fatal("expected false when remote branch doesn't exist")
	}
}

func TestRemoteBranchHasUnmergedCommits_AllMerged(t *testing.T) {
	localRepo, _ := initRepoWithRemote(t)

	// Create feature branch and push it
	if err := CreateBranch(localRepo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "push", "-u", "origin", "feature/test"); err != nil {
		t.Fatalf("git push: %v", err)
	}

	// Feature branch is at same point as main, so no unmerged commits
	hasUnmerged, err := RemoteBranchHasUnmergedCommits(localRepo, "feature/test", "main")
	if err != nil {
		t.Fatalf("RemoteBranchHasUnmergedCommits: %v", err)
	}
	if hasUnmerged {
		t.Fatal("expected false when all commits are merged")
	}
}

func TestRemoteBranchHasUnmergedCommits_HasUnmerged(t *testing.T) {
	localRepo, _ := initRepoWithRemote(t)

	// Create feature branch with an actual file change
	if err := CreateBranch(localRepo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "checkout", "feature/test"); err != nil {
		t.Fatalf("git checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localRepo, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "add", "feature.txt"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "commit", "-m", "feature work"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Push feature branch to origin
	if _, err := runGit(context.Background(), localRepo, "push", "-u", "origin", "feature/test"); err != nil {
		t.Fatalf("git push: %v", err)
	}

	// Fetch to update remote tracking refs
	if _, err := runGit(context.Background(), localRepo, "fetch", "origin"); err != nil {
		t.Fatalf("git fetch: %v", err)
	}

	// Remote branch has commit not in main
	hasUnmerged, err := RemoteBranchHasUnmergedCommits(localRepo, "feature/test", "main")
	if err != nil {
		t.Fatalf("RemoteBranchHasUnmergedCommits: %v", err)
	}
	if !hasUnmerged {
		t.Fatal("expected true when remote has unmerged commits")
	}
}
