# CCW (Claude Code Workspace)

CCW is a fast CLI that spins up Claude Code workspaces using git worktrees and tmux. It gives you a repeatable two-pane layout (Claude Code on the left, lazygit on the right), automatic worktree/branch wiring, and one-command attach/resume so you can hop between features without manual setup.

## Features
- `ccw new <repo> <branch>`: create a worktree, tmux session, and launch Claude/lazygit.
- `ccw open <workspace>`: attach to an existing workspace (recreates the session if needed).
- `ccw ls`: list workspaces with session status.
- `ccw info <workspace>`: show detailed workspace metadata.
- `ccw rm <workspace>`: best-effort cleanup of sessions, worktrees, and branches.
- `ccw stale`: list merged branches (optionally delete them).
- `ccw config`: view or set configuration values.
- `ccw completion <shell>`: generate bash/zsh/fish completions.

Workspaces live under `~/.ccw` by default (`CCW_HOME` overrides). Repos are resolved from `repos_dir` in config (default `~/github`).

## Install

### Homebrew (macOS/Linux)
```bash
brew tap justanotheratom/ccw
brew install ccw
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

## Usage (common flow)
```bash
ccw new myrepo feature/cool-thing
ccw open myrepo/feature/cool-thing
ccw ls --all
ccw info cool-thing           # fuzzy match
ccw stale --rm --force
ccw config repos_dir ~/projects
```

## Dependencies
- Required: `git`, `tmux`
- Recommended: `lazygit` (right pane)
- Claude Code CLI: installed separately; `ccw new` launches it in the left pane.

## Completions
```bash
ccw completion bash > /usr/local/etc/bash_completion.d/ccw
ccw completion zsh > "${fpath[1]}/_ccw"
ccw completion fish > ~/.config/fish/completions/ccw.fish
```

## Homebrew
The tap formula lives at `packaging/homebrew/ccw.rb` and is published to `justanotheratom/homebrew-ccw`. Update URLs/SHAs on each release.

## Release
- Bump `cmd/root.go` version, then `go test ./...`.
- Tag and push (example): `git commit -am "Release vX.Y.Z" && git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin HEAD:master --tags`
- Wait for the GitHub Actions release to finish (builds macOS/Linux amd64/arm64 and uploads assets).
- Tap automation: on `release` published, `.github/workflows/update-tap.yml` downloads assets, computes shas, rewrites the tap formula, and pushes to `justanotheratom/homebrew-ccw`. Requires `TAP_PAT` secret with `repo` access to that tap.
- Verify: `brew update && brew reinstall ccw && ccw version` (should print the new version).

## Testing
Tests rely on `git` and `tmux`. CI installs tmux; local runs may skip dependency checks by setting `CCW_SKIP_DEPS=1` when needed. Use `make test` to run the suite.

## Troubleshooting
- **"tmux session not found"** after reboot: run `ccw open <workspace>` to recreate the session.
- **Workspace shows as dead but files exist**: `ccw open <workspace>` to rebuild panes.
- **Branch not merged** when removing: merge first, or use `ccw rm --force`/`--keep-branch`.
- **Missing optional tools**: install `lazygit` for right-pane UI; `ccw` will continue without it.
- **Dependency checks in CI**: set `CCW_SKIP_DEPS=1` to skip required/optional binary checks.
