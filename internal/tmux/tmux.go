package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/term"
)

type Runner struct {
	// CCMode controls whether tmux commands executed in this process include -CC.
	// PreferCC tracks the user's preference (from config) so we can honor it when
	// spawning new macOS windows (even if CCMode is disabled in the current shell).
	CCMode   bool
	PreferCC bool
}

var (
	ErrSessionExists  = errors.New("tmux session already exists")
	ErrSessionMissing = errors.New("tmux session not found")
)

func NewRunner(ccMode bool) Runner {
	preferCC := ccMode
	// Disable -CC unless running in iTerm with a real TTY.
	if ccMode {
		if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
			ccMode = false
		}
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			ccMode = false
		}
	}
	return Runner{CCMode: ccMode, PreferCC: preferCC}
}

func (r Runner) cmdArgs(args []string) []string {
	if r.CCMode {
		return append([]string{"-CC"}, args...)
	}
	return args
}

func (r Runner) run(ctx context.Context, args ...string) (string, error) {
	fullArgs := r.cmdArgs(args)
	cmd := exec.CommandContext(ctx, "tmux", fullArgs...)
	cmd.Stdin = os.Stdin
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tmux %s: %w (stderr: %s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r Runner) SessionExists(name string) (bool, error) {
	_, err := r.run(context.Background(), "has-session", "-t", name)
	if err == nil {
		return true, nil
	}

	if code, ok := exitCode(err); ok && code == 1 {
		return false, nil
	}
	return false, err
}

func (r Runner) CreateSession(name, path string, detached bool) error {
	if exists, err := r.SessionExists(name); err != nil {
		return err
	} else if exists {
		return ErrSessionExists
	}

	args := []string{"new-session"}
	if detached {
		args = append(args, "-d")
	}
	args = append(args, "-s", name)
	if path != "" {
		args = append(args, "-c", path)
	}

	_, err := r.run(context.Background(), args...)
	return err
}

func (r Runner) KillSession(name string) error {
	_, err := r.run(context.Background(), "kill-session", "-t", name)
	if err != nil {
		if code, ok := exitCode(err); ok && code == 1 {
			return ErrSessionMissing
		}
	}
	return err
}

func (r Runner) AttachSession(name string) error {
	if runtime.GOOS == "darwin" {
		if err := openNewMacTerminalWindow(name, r.PreferCC); err == nil {
			return nil
		}
		return fmt.Errorf("failed to open macOS terminal window for tmux session %s", name)
	}

	if os.Getenv("TMUX") != "" {
		return fmt.Errorf("inside tmux; run ccw open from a non-tmux shell")
	}

	_, err := r.run(context.Background(), "attach", "-t", name)
	return err
}

func (r Runner) SplitPane(session string, horizontal bool, path string) error {
	target := normalizeTarget(session)
	args := []string{"split-window", "-t", target}
	if horizontal {
		args = append(args, "-h")
	} else {
		args = append(args, "-v")
	}
	if path != "" {
		args = append(args, "-c", path)
	}

	_, err := r.run(context.Background(), args...)
	return err
}

func (r Runner) SendKeys(target string, keys []string, enter bool) error {
	target = normalizeTarget(target)
	args := []string{"send-keys", "-t", target}
	args = append(args, keys...)
	if enter {
		args = append(args, "Enter")
	}

	_, err := r.run(context.Background(), args...)
	return err
}

func (r Runner) ListPanes(session string) (int, error) {
	target := normalizeTarget(session)
	out, err := r.run(context.Background(), "list-panes", "-t", target)
	if err != nil {
		return 0, err
	}
	if out == "" {
		return 0, nil
	}
	return len(strings.Split(out, "\n")), nil
}

func exitCode(err error) (int, bool) {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), true
	}
	return 0, false
}

func normalizeTarget(target string) string {
	if strings.ContainsAny(target, ":.") {
		return target
	}
	return target + ":"
}

func openNewMacTerminalWindow(session string, ccMode bool) error {
	tmuxBin := tmuxBinary()
	tmuxEnv := os.Getenv("TMUX")

	app := pickMacTerminalApp()
	useCC := ccMode && app == "iTerm"
	command := tmuxAttachCommand(tmuxBin, session, useCC, tmuxEnv)

	var script string
	if app == "iTerm" {
		script = fmt.Sprintf(`tell application "iTerm"
  set newWindow to (create window with default profile command "%s")
  try
    tell application "Finder"
      set screenBounds to bounds of window of desktop
    end tell
    set bounds of newWindow to screenBounds
    if %t then
      set miniaturized of newWindow to true
    end if
  end try
  activate
end tell`, command, useCC)
	} else {
		script = fmt.Sprintf(`tell application "Terminal"
  do script "%s"
  activate
end tell`, command)
	}

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func pickMacTerminalApp() string {
	paths := []string{
		"/Applications/iTerm.app",
		filepath.Join(os.Getenv("HOME"), "Applications/iTerm.app"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return "iTerm"
		}
	}
	return "Terminal"
}

func tmuxAttachCommand(tmuxBin, session string, ccMode bool, tmuxEnv string) string {
	base := fmt.Sprintf("%s attach -t %s", shellQuote(tmuxBin), session)
	if ccMode {
		base = fmt.Sprintf("%s -CC attach -t %s", shellQuote(tmuxBin), session)
	}
	if tmuxEnv == "" {
		return base
	}
	escaped := strings.ReplaceAll(tmuxEnv, `"`, `\"`)
	return fmt.Sprintf("TMUX=\"%s\" %s", escaped, base)
}

func tmuxBinary() string {
	if p, err := exec.LookPath("tmux"); err == nil {
		return p
	}
	return "tmux"
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " '\"\\$") {
		return s
	}
	return "'" + strings.ReplaceAll(s, `'`, `'\''`) + "'"
}
