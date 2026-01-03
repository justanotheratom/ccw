package tmux

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func requireTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed")
	}
}

func newSessionName() string {
	return "ccwtest" + time.Now().Format("150405000000000")
}

func TestSessionExistsFalse(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(false)
	exists, err := runner.SessionExists("ccw-no-session")
	if err != nil {
		t.Fatalf("SessionExists: %v", err)
	}
	if exists {
		t.Fatalf("expected session to be missing")
	}
}

func TestCreateAndKillSession(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(false)
	name := newSessionName()
	dir := t.TempDir()

	if err := runner.CreateSession(name, dir, true); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer runner.KillSession(name)

	exists, err := runner.SessionExists(name)
	if err != nil || !exists {
		t.Fatalf("expected session to exist after creation")
	}

	if err := runner.KillSession(name); err != nil {
		t.Fatalf("KillSession: %v", err)
	}

	exists, err = runner.SessionExists(name)
	if err != nil {
		t.Fatalf("SessionExists: %v", err)
	}
	if exists {
		t.Fatalf("expected session to be killed")
	}
}

func TestSplitPane(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(false)
	name := newSessionName()
	dir := t.TempDir()

	if err := runner.CreateSession(name, dir, true); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer runner.KillSession(name)

	if err := runner.SplitPane(name, true, dir); err != nil {
		t.Fatalf("SplitPane: %v", err)
	}

	panes, err := runner.ListPanes(name)
	if err != nil {
		t.Fatalf("ListPanes: %v", err)
	}
	if panes != 2 {
		t.Fatalf("expected 2 panes, got %d", panes)
	}
}

func TestSendKeys(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(false)
	name := newSessionName()
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "out.txt")

	if err := runner.CreateSession(name, dir, true); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer runner.KillSession(name)

	time.Sleep(100 * time.Millisecond)

	target := name + ":0.0"
	if err := runner.SendKeys(target, []string{"echo -n hello > " + targetFile}, true); err != nil {
		t.Fatalf("SendKeys: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected output to be written, got %q", string(data))
	}
}

func TestCCModeHasSession(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(true)
	_, err := runner.SessionExists("unlikely-session")
	if err != nil {
		t.Fatalf("SessionExists with CC mode: %v", err)
	}
}

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple session", "my-session", "my-session:"},
		{"session with dot", "elicited.blog--branch", "elicited.blog--branch:"},
		{"multiple dots", "foo.bar.baz--feature", "foo.bar.baz--feature:"},
		{"already has colon", "session:window", "session:window"},
		{"full target", "session:window.pane", "session:window.pane"},
		{"colon only", "session:", "session:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTarget(tt.input)
			if got != tt.expect {
				t.Errorf("normalizeTarget(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestSplitPaneWithUnderscoreInName(t *testing.T) {
	requireTmux(t)
	runner := NewRunner(false)
	// Session name with underscores (dots are converted to underscores by SafeName)
	name := "test_dotted" + time.Now().Format("150405")
	dir := t.TempDir()

	if err := runner.CreateSession(name, dir, true); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer runner.KillSession(name)

	if err := runner.SplitPane(name, true, dir); err != nil {
		t.Fatalf("SplitPane with underscored session name: %v", err)
	}

	panes, err := runner.ListPanes(name)
	if err != nil {
		t.Fatalf("ListPanes: %v", err)
	}
	if panes != 2 {
		t.Fatalf("expected 2 panes, got %d", panes)
	}
}
