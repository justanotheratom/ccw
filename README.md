# CCW - Claude Code Workspace Manager

A macOS menubar app for managing Claude Code workspaces with git worktrees and tmux sessions.

**Requirements:** macOS 14 (Sonoma) or later

## Installation

### Homebrew (Recommended)
```bash
brew install justanotheratom/ccw/ccw
```

### Direct Download
Download the latest `CCW.dmg` from [GitHub Releases](https://github.com/justanotheratom/ccw/releases) and drag `CCW.app` to `/Applications`.

## How It Works

CCW provides both a menubar GUI and a CLI:
- **Menubar App** - SwiftUI app for quick workspace access
- **CLI (`ccw`)** - Embedded Go binary for workspace management

The CLI is bundled inside the app at `CCW.app/Contents/MacOS/ccw` and symlinked to your PATH by Homebrew.

### CLI Commands
```bash
ccw ls              # List workspaces
ccw new <branch>    # Create workspace
ccw open <name>     # Open workspace
ccw rm <name>       # Remove workspace
ccw version         # Show version
```

## Project Architecture

```
CCW/
├── CCWMenubar.xcworkspace/      # Open this file in Xcode
├── CCWMenubar.xcodeproj/        # App shell project
├── CCWMenubar/                  # App target (minimal)
│   ├── Assets.xcassets/         # App icons and colors
│   └── CCWMenubarApp.swift      # App entry point
├── CCWMenubarPackage/           # Feature code (SPM package)
│   ├── Sources/CCWMenubarFeature/
│   └── Tests/CCWMenubarFeatureTests/
├── Config/                      # Build settings
│   ├── Shared.xcconfig          # Bundle ID, versions
│   ├── Debug.xcconfig
│   ├── Release.xcconfig
│   └── CCWMenubar.entitlements  # App capabilities
└── cmd/, internal/              # Go CLI source
```

## Development

### Prerequisites
- **Xcode 16+** - For building the app
- **Go 1.21+** - For building the CLI
- **iTerm2** (optional) - For terminal integration; falls back to Terminal.app

### Local Testing
```bash
# Build and launch the app
scripts/dev-menubar.sh --release

# Stream logs after launch
scripts/dev-menubar.sh --release --logs

# Skip Xcode build (relaunch only)
scripts/dev-menubar.sh --no-build
```

### Logging
The app uses structured log prefixes:

| Prefix | Component |
|--------|-----------|
| `CCWMenubar[ui]` | UI lifecycle |
| `CCWMenubar[app-state]` | AppState, workspace ops |
| `CCWMenubar[cli]` | CLI execution |

Stream all logs:
```bash
log stream --style compact \
  --predicate 'process == "CCW" || subsystem == "com.justanotheratom.ccw"' \
  --info --level info
```

## Releasing a New Version

### Step 1: Bump Version Numbers
Update version in two places:
```bash
# CLI version (cmd/root.go)
version = "X.Y.Z"

# App version (Config/Shared.xcconfig)
MARKETING_VERSION = X.Y.Z
```

### Step 2: Commit and Tag
```bash
git add cmd/root.go Config/Shared.xcconfig
git commit -m "Bump version to X.Y.Z"
git push

git tag vX.Y.Z
git push origin vX.Y.Z
```

### Step 3: Automated Release
Pushing the tag triggers GitHub Actions which:
- Builds the macOS app (universal binary)
- Signs and notarizes with Apple
- Creates DMG and uploads to GitHub Release
- Updates the Homebrew tap

### Step 4: Verify
```bash
gh release view vX.Y.Z
brew update && brew upgrade ccw
```

## Troubleshooting

### Menubar icon not appearing
```bash
# Check for zombie processes
pgrep -fl CCW

# Kill all instances and restart
pkill -9 -f CCW
open /Applications/CCW.app
```

### CLI not found
The CLI is symlinked by Homebrew. If missing:
```bash
# Use the embedded binary directly
/Applications/CCW.app/Contents/MacOS/ccw version
```

### Clean build
```bash
rm -rf .build/DerivedData
rm -rf CCWMenubarPackage/.build
scripts/dev-menubar.sh
```

## Notes

This project was scaffolded using [XcodeBuildMCP](https://github.com/cameroncooke/XcodeBuildMCP).
