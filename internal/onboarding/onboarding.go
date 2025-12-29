package onboarding

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ccw/ccw/internal/config"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// Onboarder handles first-run user onboarding.
type Onboarder struct {
	store  *config.Store
	stdin  io.Reader
	stdout io.Writer
}

// New creates an Onboarder with default stdin/stdout.
func New(store *config.Store) *Onboarder {
	return &Onboarder{
		store:  store,
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// NewWithIO creates an Onboarder with custom IO for testing.
func NewWithIO(store *config.Store, stdin io.Reader, stdout io.Writer) *Onboarder {
	return &Onboarder{store: store, stdin: stdin, stdout: stdout}
}

// NeedsOnboarding returns true if the user hasn't completed onboarding.
func NeedsOnboarding(cfg config.Config) bool {
	return !cfg.Onboarded
}

// Run executes the interactive onboarding flow.
func (o *Onboarder) Run() (config.Config, error) {
	if f, ok := o.stdin.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			return config.Config{}, fmt.Errorf("onboarding requires an interactive terminal")
		}
	}

	cfg := config.Default()
	scanner := bufio.NewScanner(o.stdin)

	o.printWelcome()

	// Question 1: repos_dir
	cfg.ReposDir = o.askString(scanner,
		"Where do you keep your git repositories?",
		"This is the root directory where ccw will look for repos.",
		cfg.ReposDir)

	// Question 2: layout
	cfg.Layout = o.askLayout(scanner)

	// Question 3: claude_dangerously_skip_permissions
	cfg.ClaudeDangerouslySkipPerms = o.askBool(scanner,
		"Skip Claude permission prompts? (--dangerously-skip-permissions)",
		"This auto-accepts all tool use. Convenient but use with caution.",
		false)

	cfg.Onboarded = true

	if err := o.store.Save(cfg); err != nil {
		return cfg, fmt.Errorf("save config: %w", err)
	}

	o.printComplete()
	return cfg, nil
}

func (o *Onboarder) printWelcome() {
	bold := color.New(color.Bold)
	bold.Fprintln(o.stdout, "\nWelcome to CCW - Claude Code Workspace Manager!")
	fmt.Fprintln(o.stdout, "Let's set up your configuration.")
	fmt.Fprintln(o.stdout)
}

func (o *Onboarder) printComplete() {
	green := color.New(color.FgGreen, color.Bold)
	green.Fprintln(o.stdout, "\nSetup complete!")
	fmt.Fprintln(o.stdout, "Run 'ccw config' to view or change settings later.")
	fmt.Fprintln(o.stdout)
}

func (o *Onboarder) askString(scanner *bufio.Scanner, question, help, defaultVal string) string {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	cyan.Fprintf(o.stdout, "%s\n", question)
	dim.Fprintf(o.stdout, "  %s\n", help)
	fmt.Fprintf(o.stdout, "  [default: %s]: ", defaultVal)

	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			return input
		}
	}
	return defaultVal
}

func (o *Onboarder) askBool(scanner *bufio.Scanner, question, help string, defaultVal bool) bool {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	defaultStr := "n"
	if defaultVal {
		defaultStr = "y"
	}

	cyan.Fprintf(o.stdout, "%s\n", question)
	dim.Fprintf(o.stdout, "  %s\n", help)
	fmt.Fprintf(o.stdout, "  [y/n, default: %s]: ", defaultStr)

	if scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch input {
		case "y", "yes", "true", "1":
			return true
		case "n", "no", "false", "0":
			return false
		}
	}
	return defaultVal
}

func (o *Onboarder) askLayout(scanner *bufio.Scanner) config.Layout {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	cyan.Fprintln(o.stdout, "Choose your pane layout:")
	dim.Fprintln(o.stdout, "  CCW opens a tmux session with two panes side-by-side.")
	fmt.Fprintln(o.stdout, "  1) claude | lazygit  (default)")
	fmt.Fprintln(o.stdout, "  2) lazygit | claude")
	fmt.Fprint(o.stdout, "  [1/2, default: 1]: ")

	layout := config.Layout{Left: "claude", Right: "lazygit"}
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "2" {
			layout.Left = "lazygit"
			layout.Right = "claude"
		}
	}
	return layout
}
