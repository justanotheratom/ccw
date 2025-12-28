package workspace

import (
	"context"
	"sync"
	"testing"
)

func TestConcurrentWorkspaceCreationSameBranch(t *testing.T) {
	reposRoot, repoName := initRepoForManager(t)
	tmuxStub := newStubTmux()
	mgr := newManagerForTest(t, reposRoot, tmuxStub)

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	create := func() {
		defer wg.Done()
		_, err := mgr.CreateWorkspace(context.Background(), repoName, "feature/concurrent", CreateOptions{NoFetch: true, NoAttach: true})
		errs <- err
	}

	wg.Add(2)
	go create()
	go create()
	wg.Wait()
	close(errs)

	var successCount, failureCount int
	for err := range errs {
		if err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	if successCount != 1 || failureCount != 1 {
		t.Fatalf("expected one success and one failure, got %d success %d failure", successCount, failureCount)
	}

	// Registry should have exactly one entry.
	reg, err := mgr.regStore.Read(context.Background())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if len(reg.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace in registry, got %d", len(reg.Workspaces))
	}
}
