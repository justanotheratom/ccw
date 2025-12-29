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

## Installation

### Homebrew (macOS/Linux)
Two options:

1) Use the tap from GitHub (recommended once published):
```bash
brew tap justanotheratom/ccw
brew install ccw
```

2) Install from the local formula in this repo (no public tap needed):
```bash
# From the repo root (after cloning)
cp packaging/homebrew/ccw.rb /opt/homebrew/Library/Taps/justanotheratom/homebrew-ccw/Formula/ccw.rb \
  || brew tap-new justanotheratom/ccw && cp packaging/homebrew/ccw.rb /opt/homebrew/Library/Taps/justanotheratom/homebrew-ccw/Formula/ccw.rb
brew install justanotheratom/ccw/ccw
```

### Manual download (examples)
```bash
curl -L https://github.com/justanotheratom/ccw/releases/latest/download/ccw-darwin-arm64 -o /usr/local/bin/ccw
chmod +x /usr/local/bin/ccw
```
Replace the URL with the binary for your OS/arch:
- macOS Intel: `ccw-darwin-amd64`
- macOS Apple Silicon: `ccw-darwin-arm64`
- Linux Intel: `ccw-linux-amd64`
- Linux ARM: `ccw-linux-arm64`

### From source
```bash
git clone https://github.com/justanotheratom/ccw.git
cd ccw
make build
sudo cp bin/ccw /usr/local/bin/
```

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

## Troubleshooting
- **"tmux session not found"** after reboot: run `ccw open <workspace>` to recreate the session.
- **Workspace shows as dead but files exist**: `ccw open <workspace>` to rebuild panes.
- **Branch not merged** when removing: merge first, or use `ccw rm --force`/`--keep-branch`.
- **Missing optional tools**: install `lazygit` for right-pane UI; `ccw` will continue without it.
- **Dependency checks in CI**: set `CCW_SKIP_DEPS=1` to skip required/optional binary checks.
