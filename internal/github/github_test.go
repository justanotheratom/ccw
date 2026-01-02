package github

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestIsGitHubRepo_HTTPS(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "https://github.com/user/repo.git")

	client := NewClient(dir)
	if !client.IsGitHubRepo() {
		t.Fatal("expected HTTPS GitHub URL to be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_SSH(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "git@github.com:user/repo.git")

	client := NewClient(dir)
	if !client.IsGitHubRepo() {
		t.Fatal("expected SSH GitHub URL to be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_SSHWithProtocol(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "ssh://git@github.com/user/repo.git")

	client := NewClient(dir)
	if !client.IsGitHubRepo() {
		t.Fatal("expected SSH protocol GitHub URL to be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_NotGitHub(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "https://gitlab.com/user/repo.git")

	client := NewClient(dir)
	if client.IsGitHubRepo() {
		t.Fatal("expected GitLab URL to not be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_NoOrigin(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")

	client := NewClient(dir)
	if client.IsGitHubRepo() {
		t.Fatal("expected repo without origin to not be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_LocalPath(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "/path/to/local/repo")

	client := NewClient(dir)
	if client.IsGitHubRepo() {
		t.Fatal("expected local path origin to not be detected as GitHub repo")
	}
}

func TestIsGitHubRepo_Caching(t *testing.T) {
	dir := t.TempDir()
	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "remote", "add", "origin", "https://github.com/user/repo.git")

	client := NewClient(dir)

	// First call
	result1 := client.IsGitHubRepo()

	// Second call should use cache
	result2 := client.IsGitHubRepo()

	if result1 != result2 {
		t.Fatal("expected cached result to match initial result")
	}
	if !result1 {
		t.Fatal("expected GitHub repo to be detected")
	}
}

func TestIsPRMerged_NoPRFound(t *testing.T) {
	// This test requires gh to be installed and authenticated
	// Skip if gh is not available
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	// Use the current repo for testing
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// Find git root
	gitRoot := findGitRoot(cwd)
	if gitRoot == "" {
		t.Skip("not in a git repository, skipping")
	}

	client := NewClient(gitRoot)
	if !client.IsGitHubRepo() {
		t.Skip("not a GitHub repository, skipping")
	}

	// Check for a branch that definitely doesn't exist as a PR
	merged, found, err := client.IsPRMerged(context.Background(), "nonexistent-branch-xyz-12345")
	if err != nil {
		// Network errors are acceptable in tests
		t.Skipf("network error (acceptable in tests): %v", err)
	}
	if found {
		t.Fatal("expected no PR to be found for nonexistent branch")
	}
	if merged {
		t.Fatal("expected merged=false when PR not found")
	}
}

func findGitRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
