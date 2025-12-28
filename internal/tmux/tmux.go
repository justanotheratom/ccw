package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	CCMode bool
}

var (
	ErrSessionExists  = errors.New("tmux session already exists")
	ErrSessionMissing = errors.New("tmux session not found")
)

func NewRunner(ccMode bool) Runner {
	return Runner{CCMode: ccMode}
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
