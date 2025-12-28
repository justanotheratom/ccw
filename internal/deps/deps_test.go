package deps

import (
	"strings"
	"testing"
)

func TestCheckExistingDependency(t *testing.T) {
	result := Check(Dependency{Name: "go"})
	if !result.Found {
		t.Fatalf("expected go to be found in PATH")
	}
	if result.Path == "" {
		t.Fatalf("expected path to be set for found dependency")
	}
}

func TestCheckMissingDependency(t *testing.T) {
	dep := Dependency{Name: "definitely-missing-binary", InstallHint: "install it"}
	result := Check(dep)
	if result.Found {
		t.Fatalf("expected dependency to be missing")
	}
	if result.Dependency.InstallHint != dep.InstallHint {
		t.Fatalf("expected install hint to be preserved")
	}
}

func TestMissingFiltersOptional(t *testing.T) {
	results := []Result{
		{Dependency: Dependency{Name: "git"}, Found: true},
		{Dependency: Dependency{Name: "lazygit", Optional: true}, Found: false},
		{Dependency: Dependency{Name: "tmux"}, Found: false},
	}

	requiredOnly := Missing(results, false)
	if len(requiredOnly) != 1 || requiredOnly[0].Dependency.Name != "tmux" {
		t.Fatalf("expected only tmux to be missing, got %+v", requiredOnly)
	}

	withOptional := Missing(results, true)
	if len(withOptional) != 2 {
		t.Fatalf("expected two missing deps, got %d", len(withOptional))
	}
}

func TestFormatResult(t *testing.T) {
	text := FormatResult(Result{
		Dependency: Dependency{Name: "tmux", DisplayName: "tmux"},
		Found:      false,
	})
	if !strings.Contains(text, "tmux") || !strings.Contains(text, "missing") {
		t.Fatalf("unexpected format: %s", text)
	}
}
