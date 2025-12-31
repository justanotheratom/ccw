package workspace

import (
	"context"
	"testing"
)

func TestE2ECreateOpenRemove(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	id := WorkspaceID(repoName, "feature/e2e")

	if _, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/e2e", CreateOptions{NoFetch: true, NoAttach: true}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := mgr.OpenWorkspace(context.Background(), id, OpenOptions{ResumeClaude: true, FocusExisting: true}); err != nil {
		t.Fatalf("OpenWorkspace: %v", err)
	}

	if err := mgr.RemoveWorkspace(context.Background(), id, RemoveOptions{Force: true}); err != nil {
		t.Fatalf("RemoveWorkspace: %v", err)
	}
}
