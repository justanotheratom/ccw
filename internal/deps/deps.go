package deps

import (
	"fmt"
	"os/exec"
)

type Dependency struct {
	Name        string
	Optional    bool
	InstallHint string
	DisplayName string
}

type Result struct {
	Dependency Dependency
	Found      bool
	Path       string
}

func DefaultDependencies() []Dependency {
	return []Dependency{
		{
			Name:        "git",
			DisplayName: "git",
			InstallHint: "Install with `brew install git` or your package manager.",
		},
		{
			Name:        "tmux",
			DisplayName: "tmux",
			InstallHint: "Install with `brew install tmux` or your package manager.",
		},
		{
			Name:        "claude",
			DisplayName: "Claude Code CLI",
			InstallHint: "Install Claude Code CLI: https://claude.com/claude-code",
		},
		{
			Name:        "lazygit",
			DisplayName: "lazygit",
			Optional:    true,
			InstallHint: "Optional: install with `brew install lazygit`.",
		},
	}
}

func Check(dep Dependency) Result {
	path, err := exec.LookPath(dep.Name)
	if err != nil {
		return Result{Dependency: dep, Found: false}
	}
	return Result{Dependency: dep, Found: true, Path: path}
}

func CheckAll(deps []Dependency) []Result {
	results := make([]Result, 0, len(deps))
	for _, dep := range deps {
		results = append(results, Check(dep))
	}
	return results
}

func Missing(results []Result, includeOptional bool) []Result {
	var missing []Result
	for _, res := range results {
		if res.Found {
			continue
		}
		if !includeOptional && res.Dependency.Optional {
			continue
		}
		missing = append(missing, res)
	}
	return missing
}

func FormatResult(res Result) string {
	status := "missing"
	if res.Found {
		status = fmt.Sprintf("found at %s", res.Path)
	}

	if res.Dependency.DisplayName != "" {
		return fmt.Sprintf("%s: %s", res.Dependency.DisplayName, status)
	}
	return fmt.Sprintf("%s: %s", res.Dependency.Name, status)
}
