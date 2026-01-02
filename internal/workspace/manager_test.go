package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ccw/ccw/internal/git"
)

type stubTmux struct {
	sessions   map[string]bool
	failCreate bool
	failSplit  bool
}

func newStubTmux() *stubTmux {
	return &stubTmux{sessions: map[string]bool{}}
}

func (s *stubTmux) SessionExists(name string) (bool, error) {
	return s.sessions[name], nil
}

func (s *stubTmux) CreateSession(name, path string, detached bool) error {
	if s.failCreate {
		return fmt.Errorf("create session fail")
	}
	s.sessions[name] = true
	return nil
}

func (s *stubTmux) KillSession(name string) error {
	delete(s.sessions, name)
	return nil
}

func (s *stubTmux) AttachSession(name string) error {
	return nil
}

func (s *stubTmux) SplitPane(session string, horizontal bool, path string) error {
	if s.failSplit {
		return fmt.Errorf("split fail")
	}
	if !s.sessions[session] {
		return fmt.Errorf("session missing")
	}
	return nil
}

func (s *stubTmux) SendKeys(target string, keys []string, enter bool) error {
	return nil
}

func (s *stubTmux) HasAttachedClients(session string) (bool, error) {
	return false, nil
}

func initRepoForManager(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	repoName := "demo"
	dir := filepath.Join(root, repoName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a bare repo to act as origin
	bareDir := filepath.Join(root, "origin.git")
	runGitCmd(t, root, "init", "--bare", "origin.git")

	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "checkout", "-b", "main")
	runGitCmd(t, dir, "config", "user.email", "test@example.com")
	runGitCmd(t, dir, "config", "user.name", "Test User")
	runGitCmd(t, dir, "remote", "add", "origin", bareDir)
	runGitCmd(t, dir, "commit", "--allow-empty", "-m", "init")
	runGitCmd(t, dir, "push", "-u", "origin", "main")
	return root, repoName
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(out))
	}
}

func newManagerForTest(t *testing.T, reposRoot string, tmuxRunner *stubTmux) *Manager {
	t.Helper()
	root := t.TempDir()
	mgr, err := NewManager(root, tmuxRunner)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	mgr.cfg.ReposDir = reposRoot
	mgr.skipDeps = true
	mgr.skipGitHubCheck = true
	if err := mgr.cfgStore.Save(mgr.cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	return mgr
}

func TestCreateWorkspaceRegistersResources(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	repoPath := filepath.Join(reposRoot, repoName)

	exists, err := git.BranchExists(repoPath, "feature/test")
	if err != nil || !exists {
		t.Fatalf("expected branch to exist: %v", err)
	}

	if _, err := os.Stat(ws.WorktreePath); err != nil {
		t.Fatalf("expected worktree path to exist: %v", err)
	}

	reg, err := mgr.regStore.Read(context.Background())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if _, ok := reg.Workspaces[WorkspaceID(repoName, "feature/test")]; !ok {
		t.Fatalf("workspace not registered")
	}

	if !tmuxStub.sessions[ws.TmuxSession] {
		t.Fatalf("tmux session not created")
	}
}

func TestCreateWorkspaceRollbackOnTmuxFailure(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	tmuxStub.failSplit = true
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	_, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err == nil {
		t.Fatalf("expected error from tmux split")
	}

	repoPath := filepath.Join(reposRoot, repoName)

	exists, err := git.BranchExists(repoPath, "feature/test")
	if err != nil {
		t.Fatalf("BranchExists: %v", err)
	}
	if exists {
		t.Fatalf("branch should be removed on rollback")
	}

	safeName := SafeName(repoName, "feature/test")
	worktreePath, _ := git.DefaultWorktreePath(filepath.Join(mgr.root, "worktrees"), safeName)
	if _, err := os.Stat(worktreePath); err == nil {
		t.Fatalf("expected worktree to be removed on rollback")
	}
}

func TestOpenWorkspaceCreatesSessionWhenMissing(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	delete(tmuxStub.sessions, ws.TmuxSession)

	if err := mgr.OpenWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), OpenOptions{ResumeClaude: true, FocusExisting: true}); err != nil {
		t.Fatalf("OpenWorkspace: %v", err)
	}

	if !tmuxStub.sessions[ws.TmuxSession] {
		t.Fatalf("expected tmux session to be recreated")
	}
}

func TestRemoveWorkspaceRemovesResources(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := mgr.RemoveWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), RemoveOptions{Force: true}); err != nil {
		t.Fatalf("RemoveWorkspace: %v", err)
	}

	repoPath := filepath.Join(reposRoot, repoName)

	exists, err := git.BranchExists(repoPath, "feature/test")
	if err != nil {
		t.Fatalf("BranchExists: %v", err)
	}
	if exists {
		t.Fatalf("expected branch to be deleted")
	}

	if _, err := os.Stat(ws.WorktreePath); err == nil {
		t.Fatalf("expected worktree path to be removed")
	}

	reg, err := mgr.regStore.Read(context.Background())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if _, ok := reg.Workspaces[WorkspaceID(repoName, "feature/test")]; ok {
		t.Fatalf("expected registry entry to be removed")
	}
}

func TestWorkspaceInfoPartialMatch(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	info, err := mgr.WorkspaceInfo(context.Background(), "feature")
	if err != nil {
		t.Fatalf("WorkspaceInfo: %v", err)
	}
	if info.ID != WorkspaceID(repoName, "feature/test") {
		t.Fatalf("unexpected workspace resolved: %s", info.ID)
	}
}

func TestStaleWorkspacesDetectsMerged(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	repoPath := filepath.Join(reposRoot, repoName)
	worktreePath, _ := git.DefaultWorktreePath(filepath.Join(mgr.root, "worktrees"), SafeName(repoName, "feature/test"))
	if err := os.WriteFile(filepath.Join(worktreePath, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGitCmd(t, worktreePath, "add", ".")
	runGitCmd(t, worktreePath, "commit", "-m", "feature work")
	runGitCmd(t, worktreePath, "push", "origin", "feature/test")
	runGitCmd(t, repoPath, "merge", "--no-ff", "feature/test")
	runGitCmd(t, repoPath, "push", "origin", "main")

	stale, err := mgr.StaleWorkspaces(context.Background(), false)
	if err != nil {
		t.Fatalf("StaleWorkspaces: %v", err)
	}
	if len(stale) != 1 || stale[0].ID != WorkspaceID(repoName, "feature/test") {
		t.Fatalf("expected one stale workspace, got %+v", stale)
	}
}

func TestSetConfigValue(t *testing.T) {
	reposRoot, _ := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	cfg, err := mgr.SetConfigValue("layout.right", "custom")
	if err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}
	if cfg.Layout.Right != "custom" {
		t.Fatalf("expected layout.right to change")
	}

	cfg, err = mgr.SetConfigValue("claude_dangerously_skip_permissions", "true")
	if err != nil {
		t.Fatalf("SetConfigValue: %v", err)
	}
	if !cfg.ClaudeDangerouslySkipPerms {
		t.Fatalf("expected claude_dangerously_skip_permissions to be true")
	}
}

func TestCreateWorkspacePathTraversalValidation(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	cases := []struct {
		name    string
		repo    string
		branch  string
		wantErr bool
	}{
		{"valid", repoName, "feature/test", false},
		{"repo traversal", "../../etc", "branch", true},
		{"branch traversal", repoName, "../../passwd", true},
		{"absolute repo", "/etc/passwd", "branch", true},
		{"absolute branch", repoName, "/etc/shadow", true},
		{"null byte", repoName + "\x00", "branch", true},
		{"backslash", repoName + "\\", "branch", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mgr.CreateWorkspace(context.Background(), tc.repo, tc.branch, CreateOptions{NoFetch: true, NoAttach: true})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for case %s", tc.name)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLookupWorkspace_ByIndex(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	// Create two workspaces
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/a", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/b", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Lookup by index 1 (should be demo/feature/a since sorted alphabetically)
	info, err := mgr.WorkspaceInfo(context.Background(), "1")
	if err != nil {
		t.Fatalf("WorkspaceInfo: %v", err)
	}
	if info.ID != WorkspaceID(repoName, "feature/a") {
		t.Fatalf("expected demo/feature/a, got %s", info.ID)
	}

	// Lookup by index 2 (should be demo/feature/b)
	info, err = mgr.WorkspaceInfo(context.Background(), "2")
	if err != nil {
		t.Fatalf("WorkspaceInfo: %v", err)
	}
	if info.ID != WorkspaceID(repoName, "feature/b") {
		t.Fatalf("expected demo/feature/b, got %s", info.ID)
	}
}

func TestLookupWorkspace_IndexOutOfRange(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	// Create one workspace
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Index 2 should be out of range
	_, err := mgr.WorkspaceInfo(context.Background(), "2")
	if err == nil {
		t.Fatal("expected error for out of range index")
	}
}

func TestLookupWorkspace_IndexZero(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	// Create one workspace
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Index 0 should fail (1-based indexing)
	_, err := mgr.WorkspaceInfo(context.Background(), "0")
	if err == nil {
		t.Fatal("expected error for index 0")
	}
}

func TestLookupWorkspace_NameTakesPrecedence(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	// Create workspaces - one with a numeric-looking name (via ID, though branch names can't be just numbers)
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/a", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/b", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Lookup by exact name should work
	info, err := mgr.WorkspaceInfo(context.Background(), WorkspaceID(repoName, "feature/b"))
	if err != nil {
		t.Fatalf("WorkspaceInfo: %v", err)
	}
	if info.ID != WorkspaceID(repoName, "feature/b") {
		t.Fatalf("expected exact name match, got %s", info.ID)
	}
}

func TestRemoveWorkspace_DeletesRemoteBranch(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Push the branch to remote
	runGitCmd(t, ws.WorktreePath, "push", "-u", "origin", "feature/test")

	repoPath := filepath.Join(reposRoot, repoName)

	// Verify remote branch exists
	exists, err := git.RemoteBranchExists(repoPath, "origin", "feature/test")
	if err != nil {
		t.Fatalf("RemoteBranchExists: %v", err)
	}
	if !exists {
		t.Fatal("expected remote branch to exist before removal")
	}

	// Remove workspace with force (to skip merge checks)
	if err := mgr.RemoveWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), RemoveOptions{Force: true}); err != nil {
		t.Fatalf("RemoveWorkspace: %v", err)
	}

	// Fetch to update remote refs
	runGitCmd(t, repoPath, "fetch", "--prune", "origin")

	// Verify remote branch is deleted
	exists, err = git.RemoteBranchExists(repoPath, "origin", "feature/test")
	if err != nil {
		t.Fatalf("RemoteBranchExists: %v", err)
	}
	if exists {
		t.Fatal("expected remote branch to be deleted")
	}
}

func TestRemoveWorkspace_SkipsRemoteIfNotExists(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	// Create workspace but don't push to remote
	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Remove should succeed even though remote branch doesn't exist
	if err := mgr.RemoveWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), RemoveOptions{Force: true}); err != nil {
		t.Fatalf("RemoveWorkspace: %v", err)
	}
}

func TestRemoveWorkspace_KeepBranchSkipsRemote(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Push the branch to remote
	runGitCmd(t, ws.WorktreePath, "push", "-u", "origin", "feature/test")

	repoPath := filepath.Join(reposRoot, repoName)

	// Remove with KeepBranch=true
	if err := mgr.RemoveWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), RemoveOptions{KeepBranch: true}); err != nil {
		t.Fatalf("RemoveWorkspace: %v", err)
	}

	// Local branch should still exist
	exists, err := git.BranchExists(repoPath, "feature/test")
	if err != nil {
		t.Fatalf("BranchExists: %v", err)
	}
	if !exists {
		t.Fatal("expected local branch to still exist with KeepBranch=true")
	}

	// Remote branch should still exist
	exists, err = git.RemoteBranchExists(repoPath, "origin", "feature/test")
	if err != nil {
		t.Fatalf("RemoteBranchExists: %v", err)
	}
	if !exists {
		t.Fatal("expected remote branch to still exist with KeepBranch=true")
	}
}

func TestRemoveWorkspace_MissingBranchStillCleansUp(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	ws, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/test", CreateOptions{NoFetch: true, NoAttach: true})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	repoPath := filepath.Join(reposRoot, repoName)

	// Remove worktree first, then delete branch to simulate a bad state
	// (workspace registered but branch manually deleted)
	runGitCmd(t, repoPath, "worktree", "remove", "--force", ws.WorktreePath)
	runGitCmd(t, repoPath, "branch", "-D", "feature/test")

	// Verify branch is gone
	exists, _ := git.BranchExists(repoPath, "feature/test")
	if exists {
		t.Fatal("expected branch to be deleted for test setup")
	}

	// Remove should still succeed and clean up registry
	if err := mgr.RemoveWorkspace(context.Background(), WorkspaceID(repoName, "feature/test"), RemoveOptions{}); err != nil {
		t.Fatalf("RemoveWorkspace should succeed even with missing branch: %v", err)
	}

	// Registry should be cleaned up
	reg, err := mgr.regStore.Read(context.Background())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if _, ok := reg.Workspaces[WorkspaceID(repoName, "feature/test")]; ok {
		t.Fatal("expected registry entry to be removed")
	}
}
