# CCW (Claude Code Workspace)

CCW is a CLI for creating and managing Claude Code development workspaces using git worktrees and tmux sessions.

## Features
- `ccw new <repo> <branch>`: create a worktree, tmux session, and launch Claude/lazygit.
- `ccw open <workspace>`: attach to an existing workspace (recreates the session if needed).
- `ccw ls`: list workspaces with session status.
- `ccw info <workspace>`: show detailed workspace metadata.
- `ccw rm <workspace>`: best-effort cleanup of sessions, worktrees, and branches.
- `ccw stale`: list merged branches (optionally delete them).
- `ccw config`: view or set configuration values.
- `ccw completion <shell>`: generate bash/zsh/fish completions.

Workspaces are stored under `~/.ccw` by default (`CCW_HOME` overrides). Repositories are resolved from `repos_dir` in config (default `~/github`).

## Usage
```bash
ccw new myrepo feature/cool-thing
ccw open myrepo/feature/cool-thing
ccw ls --all
ccw info cool-thing           # fuzzy match
ccw stale --rm --force
ccw config repos_dir ~/projects
```

## Completions
```bash
ccw completion bash > /usr/local/etc/bash_completion.d/ccw
ccw completion zsh > "${fpath[1]}/_ccw"
ccw completion fish > ~/.config/fish/completions/ccw.fish
```

## Homebrew
An example formula is provided at `packaging/homebrew/ccw.rb`. Update the tarball URL and SHA before publishing a tap.

## Release
- Tag a version (`vX.Y.Z`) to trigger the release workflow.
- Artifacts for macOS/Linux amd64 are built and uploaded from CI (`.github/workflows/release.yml`).
- Update the Homebrew formula to point at the new tag.

## Testing
Tests rely on `git` and `tmux`. CI installs tmux; local runs may skip dependency checks by setting `CCW_SKIP_DEPS=1` when needed. Use `make test` to run the suite.
