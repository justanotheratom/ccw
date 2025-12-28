package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ccw/ccw/internal/config"
	"github.com/ccw/ccw/internal/deps"
	"github.com/ccw/ccw/internal/git"
	"github.com/ccw/ccw/internal/tmux"
)

type TmuxRunner interface {
	SessionExists(name string) (bool, error)
	CreateSession(name, path string, detached bool) error
	KillSession(name string) error
	AttachSession(name string) error
	SplitPane(session string, horizontal bool, path string) error
	SendKeys(target string, keys []string, enter bool) error
}

type Manager struct {
	root     string
	cfg      config.Config
	cfgStore *config.Store
	regStore *Store
	tmux     TmuxRunner

	lazygitAvailable bool
	skipDeps         bool
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
}

type WorkspaceStatus struct {
	ID           string
	Workspace    Workspace
	SessionAlive bool
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
	if err := m.checkDepsByName("git", "tmux", "claude"); err != nil {
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

	baseBranch := opts.BaseBranch
	if baseBranch == "" {
		baseBranch = m.cfg.DefaultBase
	}

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

	if err := git.CreateWorktree(repoPath, worktreePath, branch); err != nil {
		rb.Run()
		return Workspace{}, err
	}
	rb.Add(func() { _ = git.RemoveWorktree(repoPath, worktreePath, true) })

	if err := m.bootstrapSession(safeName, worktreePath, false); err != nil {
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

	if !opts.NoAttach {
		if err := m.tmux.AttachSession(safeName); err != nil {
			return ws, err
		}
	}

	return ws, nil
}

func (m *Manager) bootstrapSession(name, path string, resume bool) error {
	if err := m.tmux.CreateSession(name, path, true); err != nil {
		return err
	}

	if err := m.tmux.SplitPane(name, true, path); err != nil {
		return err
	}

	claudeCmd := "claude"
	if resume {
		claudeCmd = fmt.Sprintf("claude --resume %s", name)
	}
	if err := m.tmux.SendKeys(name+":0.0", []string{claudeCmd}, true); err != nil {
		return err
	}

	if m.lazygitAvailable {
		_ = m.tmux.SendKeys(name+":0.1", []string{"lazygit"}, true)
	}

	return nil
}

func (m *Manager) OpenWorkspace(ctx context.Context, id string, resumeClaude bool) error {
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
		if err := m.bootstrapSession(ws.TmuxSession, ws.WorktreePath, resumeClaude); err != nil {
			return err
		}
	}

	if err := m.updateLastAccessed(ctx, resolvedID); err != nil {
		return err
	}

	return m.tmux.AttachSession(ws.TmuxSession)
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
		statuses = append(statuses, WorkspaceStatus{
			ID:           id,
			Workspace:    ws,
			SessionAlive: alive,
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

	var errs []error

	if err := m.tmux.KillSession(ws.TmuxSession); err != nil && !errors.Is(err, tmux.ErrSessionMissing) {
		errs = append(errs, fmt.Errorf("kill session: %w", err))
	}

	if !opts.KeepWorktree {
		if err := git.RemoveWorktree(ws.RepoPath, ws.WorktreePath, true); err != nil {
			errs = append(errs, fmt.Errorf("remove worktree: %w", err))
		}
	}

	if !opts.KeepBranch {
		if !opts.Force {
			merged, err := git.IsMerged(ws.RepoPath, ws.Branch, ws.BaseBranch, true)
			if err != nil {
				return err
			}
			if !merged {
				return git.ErrBranchNotMerged
			}

			unpushed, err := git.HasUnpushedCommits(ws.RepoPath, ws.Branch)
			if err != nil {
				return err
			}
			if unpushed {
				return fmt.Errorf("branch has unpushed commits")
			}
		}

		if err := git.DeleteBranch(ws.RepoPath, ws.Branch, opts.Force); err != nil {
			errs = append(errs, fmt.Errorf("delete branch: %w", err))
		}
	}

	if err := m.regStore.Update(ctx, func(reg *Registry) error {
		reg.Remove(resolvedID)
		return nil
	}); err != nil {
		errs = append(errs, fmt.Errorf("update registry: %w", err))
	}

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

	if ws, ok := reg.Workspaces[query]; ok {
		return query, ws, nil
	}

	matches := reg.FindByPartialName(query)
	if len(matches) == 1 {
		ws := reg.Workspaces[matches[0]]
		return matches[0], ws, nil
	}

	if len(matches) > 1 {
		return "", Workspace{}, fmt.Errorf("multiple workspaces match: %s", strings.Join(matches, ", "))
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

	return WorkspaceStatus{
		ID:           id,
		Workspace:    ws,
		SessionAlive: alive,
	}, nil
}

func (m *Manager) StaleWorkspaces(ctx context.Context, force bool) ([]WorkspaceStatus, error) {
	if err := m.checkDepsByName("git"); err != nil {
		return nil, err
	}

	reg, err := m.regStore.Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []WorkspaceStatus
	for id, ws := range reg.Workspaces {
		merged, err := git.IsMerged(ws.RepoPath, ws.Branch, ws.BaseBranch, false)
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
			results = append(results, WorkspaceStatus{
				ID:           id,
				Workspace:    ws,
				SessionAlive: alive,
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
	case "default_base":
		cfg.DefaultBase = value
	case "iterm_cc_mode":
		cfg.ITermCCMode = strings.ToLower(value) == "true"
	case "claude_rename_delay":
		delay, err := strconv.Atoi(value)
		if err != nil {
			return cfg, fmt.Errorf("invalid claude_rename_delay: %w", err)
		}
		cfg.ClaudeRenameDelay = delay
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
