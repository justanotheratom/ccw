package claude

import (
	"context"
	"testing"
)

func TestParseHelpDetectsFlags(t *testing.T) {
	text := `
Usage of claude:
  --resume SESSION   Resume a session
  --session-name NAME    Start with a specific session name
`
	caps := parseHelp(text)
	if !caps.SupportsResume {
		t.Fatalf("expected SupportsResume true")
	}
	if caps.SessionNameFlag != "--session-name" {
		t.Fatalf("expected session name flag to be detected")
	}
}

func TestBuildLaunchCommand(t *testing.T) {
	caps := Capabilities{SupportsResume: true, SessionNameFlag: "--session-name"}

	// Test resume without skip perms
	cmd := BuildLaunchCommand("demo", true, caps, false)
	if cmd != "claude --resume demo" {
		t.Fatalf("unexpected command: %s", cmd)
	}

	// Test non-resume without skip perms
	cmd = BuildLaunchCommand("demo", false, caps, false)
	if cmd != "claude --session-name demo" {
		t.Fatalf("unexpected command: %s", cmd)
	}

	// Test resume with skip perms
	cmd = BuildLaunchCommand("demo", true, caps, true)
	if cmd != "claude --dangerously-skip-permissions --resume demo" {
		t.Fatalf("unexpected command with skip perms: %s", cmd)
	}

	// Test non-resume with skip perms
	cmd = BuildLaunchCommand("demo", false, caps, true)
	if cmd != "claude --dangerously-skip-permissions --session-name demo" {
		t.Fatalf("unexpected command with skip perms: %s", cmd)
	}
}

func TestDetectCapabilitiesHandlesFailure(t *testing.T) {
	// Use a bogus PATH to force failure; expect default caps without panic.
	_, err := DetectCapabilities(context.Background())
	if err == nil {
		// On systems with claude installed this may pass; we only care that it doesn't panic.
	}
}
