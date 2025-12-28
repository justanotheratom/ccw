package workspace

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T, timeout time.Duration) *Store {
	t.Helper()
	store, err := newStore(t.TempDir(), timeout)
	if err != nil {
		t.Fatalf("newStore: %v", err)
	}
	return store
}

func TestRegistryLoadEmpty(t *testing.T) {
	store := newTestStore(t, 200*time.Millisecond)
	reg, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if reg.Version != CurrentVersion {
		t.Fatalf("expected version %d, got %d", CurrentVersion, reg.Version)
	}

	if len(reg.Workspaces) != 0 {
		t.Fatalf("expected empty registry, got %d items", len(reg.Workspaces))
	}
}

func TestRegistrySaveAndLoad(t *testing.T) {
	store := newTestStore(t, 200*time.Millisecond)

	ws := Workspace{
		Repo:         "demo",
		RepoPath:     "/tmp/demo",
		Branch:       "feature/test",
		BaseBranch:   "main",
		WorktreePath: "/tmp/worktree",
		CreatedAt:    time.Now(),
	}

	err := store.Update(context.Background(), func(reg *Registry) error {
		return reg.Add("demo/feature/test", ws)
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	reg, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	got, ok := reg.Get("demo/feature/test")
	if !ok {
		t.Fatalf("expected workspace to exist after save")
	}

	if got.Branch != ws.Branch || got.Repo != ws.Repo {
		t.Fatalf("loaded workspace mismatch: %+v", got)
	}
}

func TestRegistryRemove(t *testing.T) {
	store := newTestStore(t, 200*time.Millisecond)

	err := store.Update(context.Background(), func(reg *Registry) error {
		if err := reg.Add("demo/feature/test", Workspace{}); err != nil {
			return err
		}
		reg.Remove("demo/feature/test")
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	reg, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if _, ok := reg.Get("demo/feature/test"); ok {
		t.Fatalf("expected workspace to be removed")
	}
}

func TestRegistryFindByPartialName(t *testing.T) {
	store := newTestStore(t, 200*time.Millisecond)

	err := store.Update(context.Background(), func(reg *Registry) error {
		reg.Workspaces = map[string]Workspace{
			"demo/feature/foo": {},
			"demo/feature/bar": {},
			"other/task":       {},
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	reg, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	results := reg.FindByPartialName("feature")
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestRegistryLockBlocks(t *testing.T) {
	store := newTestStore(t, 150*time.Millisecond)

	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		_ = store.Update(context.Background(), func(reg *Registry) error {
			reg.Add("demo/feature/lock", Workspace{})
			close(started)
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		close(done)
	}()

	<-started

	err := store.Update(context.Background(), func(reg *Registry) error {
		reg.Add("demo/feature/second", Workspace{})
		return nil
	})

	if err == nil {
		t.Fatalf("expected lock contention error")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}

	<-done
}

func TestRegistryLoadExistingFile(t *testing.T) {
	store := newTestStore(t, 200*time.Millisecond)
	path := filepath.Join(store.root, registryFileName)

	reg := Registry{
		Version: CurrentVersion,
		Workspaces: map[string]Workspace{
			"demo/feature/test": {Repo: "demo"},
		},
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := os.MkdirAll(store.root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	loaded, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(loaded.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(loaded.Workspaces))
	}
}
