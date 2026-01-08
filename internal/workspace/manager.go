package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ccw/ccw/internal/claude"
	"github.com/ccw/ccw/internal/config"
	"github.com/ccw/ccw/internal/deps"
	"github.com/ccw/ccw/internal/git"
	"github.com/ccw/ccw/internal/github"
	"github.com/ccw/ccw/internal/tmux"
	"golang.org/x/term"
)

type TmuxRunner interface {
	SessionExists(name string) (bool, error)
	HasAttachedClients(session string) (bool, error)
	CreateSession(name, path string, detached bool) error
	KillSession(name string) error
	AttachSession(name string) error
	SplitPane(session string, horizontal bool, path string) error
	SendKeys(target string, keys []string, enter bool) error
}

var ErrWorkspaceAlreadyOpen = errors.New("workspace already open")

type OpenOptions struct {
	ResumeClaude  bool
	FocusExisting bool
	ForceAttach   bool
}

type Manager struct {
	root     string
	cfg      config.Config
	cfgStore *config.Store
	regStore *Store
	tmux     TmuxRunner

	lazygitAvailable bool
	skipDeps         bool
	claudeCaps       claude.Capabilities
	capsDetected     bool

	// GitHub clients keyed by repoPath
	ghClients map[string]*github.Client

	// skipGitHubCheck skips GitHub repo validation (for testing only)
	skipGitHubCheck bool
}

type CreateOptions struct {
	BaseBranch string
	NoAttach   bool
	NoFetch    bool
	Message    string
}

type RemoveOptions struct {
	Force        bool
	KeepBranch   bool
	KeepWorktree bool
	// ConfirmFunc is called when there are unmerged changes. It receives the
	// warning message and list of files that differ. Returns true to proceed
	// with deletion, false to abort. If nil, removal is aborted on conflicts.
	ConfirmFunc func(message string, files []string) bool
}

type WorkspaceStatus struct {
	ID           string
	Workspace    Workspace
	SessionAlive bool
	HasClients   bool
}

func NewManager(root string, tmuxRunner TmuxRunner) (*Manager, error) {
	if root == "" {
		if env := os.Getenv("CCW_HOME"); env != "" {
			root = env
		}
	}

	cfgStore, err := config.NewStore(root)
	if err != nil {
		return nil, err
	}
	cfg, err := cfgStore.Load()
	if err != nil {
		return nil, err
	}

	if root == "" {
		root = cfgStore.Root()
	}

	regStore, err := NewStore(root)
	if err != nil {
		return nil, err
	}

	if tmuxRunner == nil {
		tmuxRunner = tmux.NewRunner(cfg.ITermCCMode)
	}

	m := &Manager{
		root:     root,
		cfg:      cfg,
		cfgStore: cfgStore,
		regStore: regStore,
		tmux:     tmuxRunner,
	}

	m.detectOptionalDeps()
	return m, nil
}

func (m *Manager) detectOptionalDeps() {
	for _, dep := range deps.DefaultDependencies() {
		if dep.Name != "lazygit" {
			continue
		}
		res := deps.Check(dep)
		m.lazygitAvailable = res.Found
	}
}

// getGitHubClient returns a GitHub client for the given repo path.
// Returns nil if gh is not available or repo is not GitHub-hosted.
func (m *Manager) getGitHubClient(repoPath string) *github.Client {
	if m.ghClients == nil {
		m.ghClients = make(map[string]*github.Client)
	}

	if client, ok := m.ghClients[repoPath]; ok {
		return client
	}

	client := github.NewClient(repoPath)
	m.ghClients[repoPath] = client
	return client
}

// getPRChecker returns a MergeChecker function for the given repo path.
// The checker uses GitHub PR status to determine if a branch was merged.
func (m *Manager) getPRChecker(repoPath string) git.MergeChecker {
	client := m.getGitHubClient(repoPath)
	if client == nil || !client.IsGitHubRepo() {
		return nil
	}
	return client.IsPRMerged
}

func (m *Manager) checkDepsByName(names ...string) error {
	if m.skipDeps || os.Getenv("CCW_SKIP_DEPS") == "1" {
		return nil
	}

	var toCheck []deps.Dependency
	all := deps.DefaultDependencies()
	if len(names) == 0 {
		toCheck = all
	} else {
		for _, name := range names {
			for _, dep := range all {
				if dep.Name == name {
					toCheck = append(toCheck, dep)
				}
			}
		}
	}

	results := deps.CheckAll(toCheck)
	for _, res := range results {
		if res.Dependency.Name == "lazygit" {
			m.lazygitAvailable = res.Found
		}
	}

	missing := deps.Missing(results, false)
	if len(missing) == 0 {
		return nil
	}

	var hints []string
	for _, res := range missing {
		hints = append(hints, fmt.Sprintf("%s (%s)", res.Dependency.DisplayName, res.Dependency.InstallHint))
	}
	return fmt.Errorf("missing dependencies: %s", strings.Join(hints, "; "))
}

func (m *Manager) CreateWorkspace(ctx context.Context, repo, branch string, opts CreateOptions) (Workspace, error) {
	if err := m.checkDepsByName("git", "tmux", "claude", "gh"); err != nil {
		return Workspace{}, err
	}

	if err := validateName(repo); err != nil {
		return Workspace{}, fmt.Errorf("invalid repo name: %w", err)
	}
	if err := validateBranch(branch); err != nil {
		return Workspace{}, fmt.Errorf("invalid branch name: %w", err)
	}

	reposDir, err := m.cfg.ExpandedReposDir()
	if err != nil {
		return Workspace{}, err
	}

	repoPath := filepath.Join(reposDir, repo)
	if _, err := git.ValidateRepo(repoPath); err != nil {
		return Workspace{}, err
	}

	// Validate this is a GitHub-hosted repo (unless skipped for testing)
	if !m.skipGitHubCheck {
		ghClient := m.getGitHubClient(repoPath)
		if !ghClient.IsGitHubRepo() {
			return Workspace{}, fmt.Errorf("repository %q is not hosted on GitHub. ccw requires GitHub repositories.", repo)
		}

		// Check gh authentication
		if err := github.CheckAuthenticated(); err != nil {
			return Workspace{}, err
		}
	}

	baseBranch := opts.BaseBranch
	// If baseBranch is empty, git.CreateBranch will auto-detect main/master

	worktreeRoot := filepath.Join(m.root, "worktrees")
	workspaceID := WorkspaceID(repo, branch)
	safeName := SafeName(repo, branch)
	worktreePath, err := git.DefaultWorktreePath(worktreeRoot, safeName)
	if err != nil {
		return Workspace{}, err
	}

	rb := rollback{}

	if err := git.CreateBranch(repoPath, branch, baseBranch, !opts.NoFetch); err != nil {
		return Workspace{}, err
	}
	rb.Add(func() { _ = git.DeleteBranch(repoPath, branch, true) })

	if err := git.PushBranch(repoPath, branch); err != nil {
		rb.Run()
		return Workspace{}, err
	}
	rb.Add(func() { _ = git.DeleteRemoteBranch(repoPath, "origin", branch) })

	if err := git.CreateWorktree(repoPath, worktreePath, branch); err != nil {
		rb.Run()
		return Workspace{}, err
	}
	rb.Add(func() { _ = git.RemoveWorktree(repoPath, worktreePath, true) })

	// Copy .env from main repo to worktree if it exists
	envSrc := filepath.Join(repoPath, ".env")
	envDst := filepath.Join(worktreePath, ".env")
	if err := copyFileIfExists(envSrc, envDst); err != nil {
		rb.Run()
		return Workspace{}, fmt.Errorf("copy .env: %w", err)
	}

	if err := m.bootstrapSession(ctx, safeName, worktreePath, false); err != nil {
		rb.Run()
		return Workspace{}, err
	}
	rb.Add(func() { _ = m.tmux.KillSession(safeName) })

	now := time.Now().UTC()
	ws := Workspace{
		Repo:           repo,
		RepoPath:       repoPath,
		Branch:         branch,
		BaseBranch:     baseBranch,
		WorktreePath:   worktreePath,
		ClaudeSession:  safeName,
		TmuxSession:    safeName,
		CreatedAt:      now,
		LastAccessedAt: now,
	}

	if err := m.regStore.Update(ctx, func(reg *Registry) error {
		return reg.Add(workspaceID, ws)
	}); err != nil {
		rb.Run()
		return Workspace{}, err
	}

	if !opts.NoAttach && term.IsTerminal(int(os.Stdout.Fd())) {
		if err := m.tmux.AttachSession(safeName); err != nil {
			return ws, err
		}
	}

	return ws, nil
}

func (m *Manager) bootstrapSession(ctx context.Context, name, path string, resume bool) error {
	if err := m.tmux.CreateSession(name, path, true); err != nil {
		return err
	}

	if err := m.tmux.SplitPane(name, true, path); err != nil {
		return err
	}

	caps := m.claudeCapabilities(ctx)
	claudeCmd := claude.BuildLaunchCommand(name, resume, caps, m.cfg.ClaudeDangerouslySkipPerms)
	if err := m.tmux.SendKeys(name+":0.0", []string{claudeCmd}, true); err != nil {
		return err
	}

	if m.lazygitAvailable {
		_ = m.tmux.SendKeys(name+":0.1", []string{"lazygit"}, true)
	}

	return nil
}

func (m *Manager) OpenWorkspace(ctx context.Context, id string, opts OpenOptions) error {
	if err := m.checkDepsByName("git", "tmux", "claude"); err != nil {
		return err
	}

	resolvedID, ws, err := m.lookupWorkspace(ctx, id)
	if err != nil {
		return err
	}

	sessionExists, err := m.tmux.SessionExists(ws.TmuxSession)
	if err != nil {
		return err
	}

	if !sessionExists {
		if err := m.bootstrapSession(ctx, ws.TmuxSession, ws.WorktreePath, opts.ResumeClaude); err != nil {
			return err
		}
	}

	if sessionExists {
		hasClients, _ := m.tmux.HasAttachedClients(ws.TmuxSession)
		if hasClients && !opts.FocusExisting {
			return ErrWorkspaceAlreadyOpen
		}
	}

	if err := m.updateLastAccessed(ctx, resolvedID); err != nil {
		return err
	}

	if opts.ForceAttach || term.IsTerminal(int(os.Stdout.Fd())) {
		return m.tmux.AttachSession(ws.TmuxSession)
	}
	return nil
}

func (m *Manager) updateLastAccessed(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return m.regStore.Update(ctx, func(reg *Registry) error {
		ws, ok := reg.Workspaces[id]
		if !ok {
			return fmt.Errorf("workspace %s not found", id)
		}
		ws.LastAccessedAt = now
		reg.Workspaces[id] = ws
		return nil
	})
}

func (m *Manager) ListWorkspaces(ctx context.Context) ([]WorkspaceStatus, error) {
	reg, err := m.regStore.Read(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(reg.Workspaces))
	for id := range reg.Workspaces {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var statuses []WorkspaceStatus
	for _, id := range ids {
		ws := reg.Workspaces[id]
		alive, err := m.tmux.SessionExists(ws.TmuxSession)
		if err != nil {
			alive = false
		}
		hasClients := false
		if alive {
			hasClients, _ = m.tmux.HasAttachedClients(ws.TmuxSession)
		}
		statuses = append(statuses, WorkspaceStatus{
			ID:           id,
			Workspace:    ws,
			SessionAlive: alive,
			HasClients:   hasClients,
		})
	}

	return statuses, nil
}

func (m *Manager) RemoveWorkspace(ctx context.Context, id string, opts RemoveOptions) error {
	if err := m.checkDepsByName("git", "tmux"); err != nil {
		return err
	}

	if err := validateName(id); err != nil {
		return fmt.Errorf("invalid workspace identifier: %w", err)
	}

	resolvedID, ws, err := m.lookupWorkspace(ctx, id)
	if err != nil {
		return err
	}

	// Track if branch is confirmed merged (via PR check) - used to force delete local branch
	merged := false

	// Run all safety checks BEFORE any destructive actions
	if !opts.KeepBranch && !opts.Force {
		branchExists, _ := git.BranchExists(ws.RepoPath, ws.Branch)

		if branchExists {
			// Resolve base branch for error messages
			baseBranch := ws.BaseBranch
			if baseBranch == "" {
				if detected, err := git.DetectDefaultBranch(ws.RepoPath); err == nil {
					baseBranch = detected
				}
			}

			// Fetch before checking merge status
			if err := git.Fetch(ws.RepoPath, true); err != nil {
				return err
			}

			var prChecker git.MergeChecker
			if !m.skipGitHubCheck {
				// Check gh authentication before PR-based merge detection
				if err := github.CheckAuthenticated(); err != nil {
					return err
				}
				prChecker = m.getPRChecker(ws.RepoPath)
			}
			merged, err = git.IsMergedWithPR(ctx, ws.RepoPath, ws.Branch, ws.BaseBranch, false, prChecker)
			if err != nil {
				return err
			}
			if !merged {
				files, _ := git.GetDiffFiles(ws.RepoPath, ws.Branch, ws.BaseBranch)
				msg := fmt.Sprintf("Branch %q has changes not in %q", ws.Branch, baseBranch)
				if opts.ConfirmFunc != nil && opts.ConfirmFunc(msg, files) {
					opts.Force = true
				} else if opts.ConfirmFunc == nil {
					return fmt.Errorf("branch %q is not merged into %q.\nUse --force to delete anyway, or --keep-branch to only remove the workspace.", ws.Branch, baseBranch)
				} else {
					return fmt.Errorf("aborted")
				}
			}

			// Skip unpushed/unmerged checks if branch is already merged - work is safe in base branch
			if !opts.Force && !merged {
				unpushed, err := git.HasUnpushedCommits(ws.RepoPath, ws.Branch)
				if err != nil {
					return err
				}
				if unpushed {
					return fmt.Errorf("branch %q has unpushed commits. Push or use --force/--keep-branch.", ws.Branch)
				}

				remoteUnmerged, err := git.RemoteBranchHasUnmergedCommitsWithPR(ctx, ws.RepoPath, ws.Branch, ws.BaseBranch, prChecker)
				if err != nil {
					return err
				}
				if remoteUnmerged {
					files, _ := git.GetDiffFiles(ws.RepoPath, "origin/"+ws.Branch, ws.BaseBranch)
					msg := fmt.Sprintf("Remote branch %q has changes not in %q", "origin/"+ws.Branch, baseBranch)
					if opts.ConfirmFunc != nil && opts.ConfirmFunc(msg, files) {
						opts.Force = true
					} else if opts.ConfirmFunc == nil {
						return fmt.Errorf("remote branch %q has commits not merged into %q.\nUse --force to delete anyway, or --keep-branch to only remove the workspace.", "origin/"+ws.Branch, baseBranch)
					} else {
						return fmt.Errorf("aborted")
					}
				}
			}
		}
	}

	// Now perform destructive actions
	var errs []error

	if !opts.KeepWorktree {
		if err := git.RemoveWorktree(ws.RepoPath, ws.WorktreePath, true); err != nil {
			errs = append(errs, fmt.Errorf("remove worktree: %w", err))
		}
	}

	if !opts.KeepBranch {
		branchExists, _ := git.BranchExists(ws.RepoPath, ws.Branch)

		if branchExists {
			// Force delete if we verified branch is merged via PR (git -d may fail if remote is gone)
			if err := git.DeleteBranch(ws.RepoPath, ws.Branch, opts.Force || merged); err != nil && !errors.Is(err, git.ErrBranchNotFound) {
				errs = append(errs, fmt.Errorf("delete branch: %w", err))
			}
		}

		// Delete remote branch if it exists
		if exists, _ := git.RemoteBranchExists(ws.RepoPath, "origin", ws.Branch); exists {
			if err := git.DeleteRemoteBranch(ws.RepoPath, "origin", ws.Branch); err != nil {
				errs = append(errs, fmt.Errorf("delete remote branch: %w", err))
			}
		}
	}

	if err := m.regStore.Update(ctx, func(reg *Registry) error {
		reg.Remove(resolvedID)
		return nil
	}); err != nil {
		errs = append(errs, fmt.Errorf("update registry: %w", err))
	}

	// Kill tmux session LAST since ccw rm might be called from within the workspace
	if err := m.tmux.KillSession(ws.TmuxSession); err != nil && !errors.Is(err, tmux.ErrSessionMissing) {
		errs = append(errs, fmt.Errorf("kill session: %w", err))
	}

	// Close the iTerm control window if it exists (best-effort, no error on failure)
	tmux.CloseITermControlWindow(ws.TmuxSession)

	if len(errs) > 0 {
		return combineErrors(errs)
	}

	return nil
}

func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return errors.New(strings.Join(messages, "; "))
}

func (m *Manager) lookupWorkspace(ctx context.Context, query string) (string, Workspace, error) {
	if err := validateName(query); err != nil {
		return "", Workspace{}, fmt.Errorf("invalid workspace identifier: %w", err)
	}

	reg, err := m.regStore.Read(ctx)
	if err != nil {
		return "", Workspace{}, err
	}

	// First, try exact name match (takes precedence over index)
	if ws, ok := reg.Workspaces[query]; ok {
		return query, ws, nil
	}

	// If query is a positive integer, try index-based lookup
	if idx, err := strconv.Atoi(query); err == nil && idx > 0 {
		ids := make([]string, 0, len(reg.Workspaces))
		for id := range reg.Workspaces {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		if idx > len(ids) {
			return "", Workspace{}, fmt.Errorf("workspace index %d out of range (have %d workspaces)", idx, len(ids))
		}
		id := ids[idx-1] // 1-based index
		return id, reg.Workspaces[id], nil
	}

	// Fall back to partial name matching
	matches := reg.FindByPartialName(query)
	if len(matches) == 1 {
		ws := reg.Workspaces[matches[0]]
		return matches[0], ws, nil
	}

	if len(matches) > 1 {
		return "", Workspace{}, fmt.Errorf("multiple workspaces match: %s\nPlease specify a full workspace ID (repo/branch).", strings.Join(matches, ", "))
	}

	return "", Workspace{}, fmt.Errorf("workspace %s not found (try ccw ls)", query)
}

func (m *Manager) WorkspaceInfo(ctx context.Context, query string) (WorkspaceStatus, error) {
	id, ws, err := m.lookupWorkspace(ctx, query)
	if err != nil {
		return WorkspaceStatus{}, err
	}

	alive, err := m.tmux.SessionExists(ws.TmuxSession)
	if err != nil {
		alive = false
	}
	hasClients := false
	if alive {
		hasClients, _ = m.tmux.HasAttachedClients(ws.TmuxSession)
	}

	return WorkspaceStatus{
		ID:           id,
		Workspace:    ws,
		SessionAlive: alive,
		HasClients:   hasClients,
	}, nil
}

func (m *Manager) StaleWorkspaces(ctx context.Context, force bool) ([]WorkspaceStatus, error) {
	if err := m.checkDepsByName("git"); err != nil {
		return nil, err
	}

	// Check gh authentication for PR-based merge detection (unless skipped)
	if !m.skipGitHubCheck {
		if err := github.CheckAuthenticated(); err != nil {
			return nil, err
		}
	}

	reg, err := m.regStore.Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []WorkspaceStatus
	for id, ws := range reg.Workspaces {
		var prChecker git.MergeChecker
		if !m.skipGitHubCheck {
			prChecker = m.getPRChecker(ws.RepoPath)
		}
		merged, err := git.IsMergedWithPR(ctx, ws.RepoPath, ws.Branch, ws.BaseBranch, false, prChecker)
		if err != nil {
			if force {
				continue
			}
			return nil, err
		}
		if merged {
			alive, err := m.tmux.SessionExists(ws.TmuxSession)
			if err != nil {
				alive = false
			}
			hasClients := false
			if alive {
				hasClients, _ = m.tmux.HasAttachedClients(ws.TmuxSession)
			}
			results = append(results, WorkspaceStatus{
				ID:           id,
				Workspace:    ws,
				SessionAlive: alive,
				HasClients:   hasClients,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (m *Manager) GetConfig() config.Config {
	return m.cfg
}

func (m *Manager) SetConfigValue(key, value string) (config.Config, error) {
	cfg := m.cfg
	switch key {
	case "repos_dir":
		cfg.ReposDir = value
	case "iterm_cc_mode":
		cfg.ITermCCMode = strings.ToLower(value) == "true"
	case "claude_rename_delay":
		delay, err := strconv.Atoi(value)
		if err != nil {
			return cfg, fmt.Errorf("invalid claude_rename_delay: %w", err)
		}
		cfg.ClaudeRenameDelay = delay
	case "layout.left":
		cfg.Layout.Left = value
	case "layout.right":
		cfg.Layout.Right = value
	case "claude_dangerously_skip_permissions":
		cfg.ClaudeDangerouslySkipPerms = strings.ToLower(value) == "true"
	case "onboarded":
		cfg.Onboarded = strings.ToLower(value) == "true"
	default:
		return cfg, fmt.Errorf("unknown config key: %s", key)
	}

	if err := m.cfgStore.Save(cfg); err != nil {
		return cfg, err
	}
	m.cfg = cfg
	return cfg, nil
}

func (m *Manager) ResetConfig() (config.Config, error) {
	cfg := config.Default()
	if err := m.cfgStore.Save(cfg); err != nil {
		return cfg, err
	}
	m.cfg = cfg
	return cfg, nil
}

func (m *Manager) claudeCapabilities(ctx context.Context) claude.Capabilities {
	if m.capsDetected {
		return m.claudeCaps
	}

	m.capsDetected = true
	if m.skipDeps || os.Getenv("CCW_SKIP_DEPS") == "1" {
		m.claudeCaps = claude.DefaultCapabilities()
		return m.claudeCaps
	}

	caps, err := claude.DetectCapabilities(ctx)
	if err != nil {
		m.claudeCaps = claude.DefaultCapabilities()
		return m.claudeCaps
	}
	m.claudeCaps = caps
	return m.claudeCaps
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("name cannot contain '..'")
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("name cannot be an absolute path")
	}
	if strings.ContainsAny(name, "\x00") {
		return fmt.Errorf("name contains invalid characters")
	}
	if strings.ContainsAny(name, "\\") {
		return fmt.Errorf("name cannot contain backslashes")
	}
	return nil
}

func validateBranch(branch string) error {
	if err := validateName(branch); err != nil {
		return err
	}
	// Allow branch slashes but forbid traversal.
	if strings.Contains(branch, "/../") || strings.HasPrefix(branch, "../") || strings.HasSuffix(branch, "/..") {
		return fmt.Errorf("branch cannot traverse directories")
	}
	return nil
}

type rollback struct {
	steps []func()
}

func (r *rollback) Add(fn func()) {
	r.steps = append(r.steps, fn)
}

func (r *rollback) Run() {
	for i := len(r.steps) - 1; i >= 0; i-- {
		r.steps[i]()
	}
}

// copyFileIfExists copies src to dst if src exists. Returns nil if src doesn't exist.
func copyFileIfExists(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
