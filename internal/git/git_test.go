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

// Tests for IsMergedWithPR with mocked PR checker

func TestIsMergedWithPR_PRMerged(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// Add unmerged commit
	if _, err := runGit(context.Background(), repo, "checkout", "feature/test"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "add", "feature.txt"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "commit", "-m", "feature work"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if _, err := runGit(context.Background(), repo, "checkout", "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	// Without PR checker, branch should be unmerged (git heuristics)
	merged, err := IsMerged(repo, "feature/test", "main", false)
	if err != nil {
		t.Fatalf("IsMerged: %v", err)
	}
	if merged {
		t.Fatal("expected branch to be unmerged without PR checker")
	}

	// With PR checker that returns merged, branch should be merged
	prChecker := func(ctx context.Context, branch string) (bool, bool, error) {
		return true, true, nil // merged=true, found=true
	}

	merged, err = IsMergedWithPR(context.Background(), repo, "feature/test", "main", false, prChecker)
	if err != nil {
		t.Fatalf("IsMergedWithPR: %v", err)
	}
	if !merged {
		t.Fatal("expected branch to be merged with PR checker returning merged")
	}
}

func TestIsMergedWithPR_PRNotMerged(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// PR checker returns not merged
	prChecker := func(ctx context.Context, branch string) (bool, bool, error) {
		return false, true, nil // merged=false, found=true
	}

	merged, err := IsMergedWithPR(context.Background(), repo, "feature/test", "main", false, prChecker)
	if err != nil {
		t.Fatalf("IsMergedWithPR: %v", err)
	}
	if merged {
		t.Fatal("expected branch to not be merged when PR checker returns not merged")
	}
}

func TestIsMergedWithPR_NoPR_FallbackToGit(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// PR checker returns not found, should fall back to git
	prChecker := func(ctx context.Context, branch string) (bool, bool, error) {
		return false, false, nil // found=false
	}

	// Branch is at same point as main, git heuristics should say merged
	merged, err := IsMergedWithPR(context.Background(), repo, "feature/test", "main", false, prChecker)
	if err != nil {
		t.Fatalf("IsMergedWithPR: %v", err)
	}
	if !merged {
		t.Fatal("expected branch to be merged via git heuristics when no PR found")
	}
}

func TestIsMergedWithPR_NilChecker_UsesGit(t *testing.T) {
	repo := initRepo(t)
	if err := CreateBranch(repo, "feature/test", "main", false); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	// nil PR checker, should use git heuristics
	merged, err := IsMergedWithPR(context.Background(), repo, "feature/test", "main", false, nil)
	if err != nil {
		t.Fatalf("IsMergedWithPR: %v", err)
	}
	if !merged {
		t.Fatal("expected branch to be merged via git heuristics with nil PR checker")
	}
}

func TestRemoteBranchHasUnmergedCommitsWithPR_PRMerged(t *testing.T) {
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
	if _, err := runGit(context.Background(), localRepo, "push", "-u", "origin", "feature/test"); err != nil {
		t.Fatalf("git push: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "fetch", "origin"); err != nil {
		t.Fatalf("git fetch: %v", err)
	}

	// PR checker returns merged
	prChecker := func(ctx context.Context, branch string) (bool, bool, error) {
		return true, true, nil // merged=true, found=true
	}

	hasUnmerged, err := RemoteBranchHasUnmergedCommitsWithPR(context.Background(), localRepo, "feature/test", "main", prChecker)
	if err != nil {
		t.Fatalf("RemoteBranchHasUnmergedCommitsWithPR: %v", err)
	}
	if hasUnmerged {
		t.Fatal("expected no unmerged commits when PR is merged")
	}
}

// advanceRemote clones the bare remote, commits a new file on "main", pushes,
// and returns the new SHA.
func advanceRemote(t *testing.T, bareRemote string) string {
	t.Helper()

	tmpClone := t.TempDir()
	if _, err := runGit(context.Background(), tmpClone, "clone", bareRemote, "."); err != nil {
		t.Fatalf("clone bare: %v", err)
	}
	// Ensure we're on main (bare remote HEAD may default to master).
	if _, err := runGit(context.Background(), tmpClone, "checkout", "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}
	if _, err := runGit(context.Background(), tmpClone, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("config email: %v", err)
	}
	if _, err := runGit(context.Background(), tmpClone, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("config name: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpClone, "remote-change.txt"), []byte("remote"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runGit(context.Background(), tmpClone, "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runGit(context.Background(), tmpClone, "commit", "-m", "remote advance"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if _, err := runGit(context.Background(), tmpClone, "push", "origin", "main"); err != nil {
		t.Fatalf("git push: %v", err)
	}

	sha, err := runGit(context.Background(), tmpClone, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return sha
}

// --- DetectDefaultBranch tests with origin/HEAD ---

func TestDetectDefaultBranch_OriginHEAD(t *testing.T) {
	localRepo, _ := initRepoWithRemote(t)

	// Create a master branch too so both exist.
	if _, err := runGit(context.Background(), localRepo, "branch", "master"); err != nil {
		t.Fatalf("create master: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "push", "origin", "master"); err != nil {
		t.Fatalf("push master: %v", err)
	}

	// Set origin/HEAD to point to main.
	if _, err := runGit(context.Background(), localRepo, "remote", "set-head", "origin", "main"); err != nil {
		t.Fatalf("set-head: %v", err)
	}

	branch, err := DetectDefaultBranch(localRepo)
	if err != nil {
		t.Fatalf("DetectDefaultBranch: %v", err)
	}
	if branch != "main" {
		t.Fatalf("expected main (via origin/HEAD), got %s", branch)
	}
}

func TestDetectDefaultBranch_FallbackMain(t *testing.T) {
	// initRepo creates a local-only repo with main — no origin/HEAD.
	repo := initRepo(t)
	branch, err := DetectDefaultBranch(repo)
	if err != nil {
		t.Fatalf("DetectDefaultBranch: %v", err)
	}
	if branch != "main" {
		t.Fatalf("expected main, got %s", branch)
	}
}

func TestDetectDefaultBranch_FallbackMaster(t *testing.T) {
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

// --- SyncLocalBranch tests ---

func TestSyncLocalBranch_FastForward(t *testing.T) {
	localRepo, bareRemote := initRepoWithRemote(t)

	// Advance remote past local.
	newSHA := advanceRemote(t, bareRemote)

	// Fetch so origin/main is updated, but local main stays behind.
	if _, err := runGit(context.Background(), localRepo, "fetch", "origin"); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	// Local main should be behind.
	localBefore, _ := runGit(context.Background(), localRepo, "rev-parse", "main")
	if localBefore == newSHA {
		t.Fatal("local should be behind remote before sync")
	}

	// main is checked out, so this should merge --ff-only.
	if err := SyncLocalBranch(localRepo, "main"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}

	localAfter, _ := runGit(context.Background(), localRepo, "rev-parse", "main")
	if localAfter != newSHA {
		t.Fatalf("expected local main to be %s, got %s", newSHA, localAfter)
	}
}

func TestSyncLocalBranch_NotCheckedOut(t *testing.T) {
	localRepo, bareRemote := initRepoWithRemote(t)

	// Switch to a different branch so main is NOT checked out.
	if _, err := runGit(context.Background(), localRepo, "checkout", "-b", "other"); err != nil {
		t.Fatalf("checkout -b other: %v", err)
	}

	newSHA := advanceRemote(t, bareRemote)
	if _, err := runGit(context.Background(), localRepo, "fetch", "origin"); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if err := SyncLocalBranch(localRepo, "main"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}

	localAfter, _ := runGit(context.Background(), localRepo, "rev-parse", "main")
	if localAfter != newSHA {
		t.Fatalf("expected local main to be %s, got %s", newSHA, localAfter)
	}
}

func TestSyncLocalBranch_AlreadyUpToDate(t *testing.T) {
	localRepo, _ := initRepoWithRemote(t)

	shaBefore, _ := runGit(context.Background(), localRepo, "rev-parse", "main")

	if err := SyncLocalBranch(localRepo, "main"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}

	shaAfter, _ := runGit(context.Background(), localRepo, "rev-parse", "main")
	if shaBefore != shaAfter {
		t.Fatal("expected no change when already up to date")
	}
}

func TestSyncLocalBranch_Diverged(t *testing.T) {
	localRepo, bareRemote := initRepoWithRemote(t)

	// Add a local-only commit so main diverges from origin/main.
	if err := os.WriteFile(filepath.Join(localRepo, "local.txt"), []byte("local"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "add", "."); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := runGit(context.Background(), localRepo, "commit", "-m", "local commit"); err != nil {
		t.Fatalf("commit: %v", err)
	}

	localBefore, _ := runGit(context.Background(), localRepo, "rev-parse", "main")

	// Advance remote too.
	advanceRemote(t, bareRemote)
	if _, err := runGit(context.Background(), localRepo, "fetch", "origin"); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	// Should skip gracefully — local has diverged.
	if err := SyncLocalBranch(localRepo, "main"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}

	localAfter, _ := runGit(context.Background(), localRepo, "rev-parse", "main")
	if localBefore != localAfter {
		t.Fatal("expected local main to be unchanged when diverged")
	}
}

func TestSyncLocalBranch_NoRemote(t *testing.T) {
	repo := initRepo(t) // No origin remote.

	// Should be a no-op, not an error.
	if err := SyncLocalBranch(repo, "main"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}
}

func TestSyncLocalBranch_NoLocalBranch(t *testing.T) {
	localRepo, _ := initRepoWithRemote(t)

	// Try to sync a branch that doesn't exist locally.
	if err := SyncLocalBranch(localRepo, "nonexistent"); err != nil {
		t.Fatalf("SyncLocalBranch: %v", err)
	}
}
