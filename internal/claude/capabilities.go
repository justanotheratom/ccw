package claude

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type Capabilities struct {
	SupportsResume  bool
	SessionNameFlag string // e.g., "--session-name" or "--name"; empty if not supported
}

func DefaultCapabilities() Capabilities {
	return Capabilities{
		SupportsResume: true,
	}
}

// DetectCapabilities inspects `claude --help` to discover supported flags. It
// returns a best-effort guess and does not fail if detection cannot run.
func DetectCapabilities(ctx context.Context) (Capabilities, error) {
	cmd := exec.CommandContext(ctx, "claude", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return DefaultCapabilities(), err
	}

	return parseHelp(out.String()), nil
}

func parseHelp(helpText string) Capabilities {
	text := strings.ToLower(helpText)
	caps := DefaultCapabilities()

	if strings.Contains(text, "--resume") {
		caps.SupportsResume = true
	}

	if strings.Contains(text, "--session-name") {
		caps.SessionNameFlag = "--session-name"
	} else if strings.Contains(text, "--name") {
		caps.SessionNameFlag = "--name"
	}

	return caps
}

func BuildLaunchCommand(name string, resume bool, caps Capabilities) string {
	if resume && caps.SupportsResume {
		return "claude --resume " + name
	}

	if caps.SessionNameFlag != "" {
		return "claude " + caps.SessionNameFlag + " " + name
	}

	return "claude"
}
