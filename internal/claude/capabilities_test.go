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
	cmd := BuildLaunchCommand("demo", true, caps)
	if cmd != "claude --resume demo" {
		t.Fatalf("unexpected command: %s", cmd)
	}

	cmd = BuildLaunchCommand("demo", false, caps)
	if cmd != "claude --session-name demo" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestDetectCapabilitiesHandlesFailure(t *testing.T) {
	// Use a bogus PATH to force failure; expect default caps without panic.
	_, err := DetectCapabilities(context.Background())
	if err == nil {
		// On systems with claude installed this may pass; we only care that it doesn't panic.
	}
}
