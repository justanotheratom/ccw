# CCW Menubar App - Implementation Plan

## Overview

A native macOS menubar application wrapping the CCW CLI, providing quick access to workspace management.

**Decisions Made:**
- **Tech Stack**: Native Swift/SwiftUI
- **CLI Integration**: Shell out to bundled `ccw` binary (no system fallback)
- **Scope**: Full feature parity with CLI

---

## Workspace States

| State | Definition | Menubar Icon | Click Action |
|-------|------------|--------------|--------------|
| **Connected** | Tmux session exists AND has attached clients (visible in iTerm) | 🟢 Green | Focus existing iTerm window |
| **Alive** | Tmux session exists but NO attached clients (running in background) | 🟡 Yellow | Attach in new maximized iTerm |
| **Dead** | Tmux session does NOT exist | 🔴 Red | Recreate session, then attach |
| **Stale** | Workspace branch has been merged into its base branch | ⚪ Gray | Show in "Stale Workspaces" submenu |

**Detection:**
- Connected: `tmux list-clients -t <session>` returns non-empty
- Alive: `tmux has-session -t <session>` succeeds, but no clients
- Dead: `tmux has-session -t <session>` fails
- Stale: Branch merged into base (git merge-base check)

**Status Item Icon Rule:** Use the highest-priority state across workspaces (Connected > Alive > Dead). Stale only appears in submenu.

---

## Project Structure

```
CCWMenubar/
├── CCWMenubar/
│   ├── App/
│   │   ├── CCWMenubarApp.swift        # @main entry with MenuBarExtra
│   │   └── AppDelegate.swift          # NSApplicationDelegate
│   │
│   ├── Core/
│   │   ├── CLI/
│   │   │   ├── CLIBridge.swift        # Process spawning, JSON parsing
│   │   │   └── CLIError.swift         # Error types
│   │   │
│   │   ├── Models/
│   │   │   ├── Workspace.swift        # Workspace model (Codable)
│   │   │   ├── WorkspaceStatus.swift  # Status with alive flag
│   │   │   └── CCWConfig.swift        # Config model
│   │   │
│   │   └── Services/
│   │       └── WorkspaceService.swift # Business logic
│   │
│   ├── ViewModels/
│   │   ├── AppState.swift             # Main @Observable state
│   │   ├── NewWorkspaceViewModel.swift
│   │   └── SettingsViewModel.swift
│   │
│   ├── Views/
│   │   ├── MenuBar/
│   │   │   ├── MenuBarView.swift      # Main dropdown
│   │   │   ├── WorkspaceRow.swift     # Single workspace item
│   │   │   └── StatusIndicator.swift  # Alive/dead dot
│   │   │
│   │   ├── Windows/
│   │   │   ├── NewWorkspaceWindow.swift
│   │   │   ├── SettingsWindow.swift
│   │   │   ├── OnboardingWindow.swift
│   │   │   └── StaleWorkspacesWindow.swift
│   │   │
│   │   └── Shared/
│   │       └── ConfirmationDialog.swift
│   │
│   └── Resources/
│       └── Assets.xcassets/
│           └── MenubarIcon.imageset/
│
└── CCWMenubarTests/
```

---

## Key Components

### 1. CLIBridge (actor)

Thread-safe CLI execution layer:

```swift
actor CLIBridge {
    func listWorkspaces() async throws -> [WorkspaceStatus]  // ccw ls --json
    func openWorkspace(_ id: String, resume: Bool = true) async throws // ccw open <id> [--no-resume]
    func createWorkspace(repo:branch:base:) async throws     // ccw new
    func removeWorkspace(_ id: String) async throws          // ccw rm
    func staleWorkspaces() async throws -> [WorkspaceStatus] // ccw stale --json
    func workspaceInfo(_ id: String) async throws -> WorkspaceInfo // ccw info --json
    func listRepos() async throws -> [String]                // ccw repos --json
    func checkDependencies() async throws -> [String: DepStatus] // ccw check --json
    func getConfig() async throws -> CCWConfig               // read ~/.ccw/config.json
    func setConfig(key:value:) async throws                  // ccw config <k> <v>
}
```

### 2. Data Models

```swift
struct WorkspaceStatus: Codable, Identifiable {
    let id: String              // "repo/branch"
    let workspace: Workspace
    let sessionAlive: Bool
    let hasClients: Bool
}

struct Workspace: Codable {
    let repo, repoPath, branch, baseBranch: String
    let worktreePath, claudeSession, tmuxSession: String
    let createdAt, lastAccessedAt: Date
}

struct CCWConfig: Codable {
    var version: Int
    var reposDir: String
    var itermCCMode: Bool
    var claudeRenameDelay: Int
    var layout: Layout
    var onboarded: Bool
    var claudeDangerouslySkipPermissions: Bool
}

typealias WorkspaceInfo = WorkspaceStatus

struct DepStatus: Codable {
    let installed: Bool
    let path: String
    let optional: Bool?
}
```

### 3. AppState (@MainActor)

Centralized state management:

```swift
@MainActor class AppState: ObservableObject {
    @Published var workspaces: [WorkspaceStatus] = []
    @Published var isLoading = false
    @Published var needsOnboarding = false

    func refreshWorkspaces() async  // Called when menu opens
    func openWorkspace(_ id: String) async
    func createWorkspace(repo:branch:base:) async
    func removeWorkspace(_ id: String) async
}
```

**Refresh Strategy**: Fetch workspace list on-demand when user clicks menubar icon (no background polling needed). Use a menu-open hook so refresh runs every open.

---

## UI Mockups

### Main Menu Dropdown
```
┌──────────────────────────────────────────┐
│ CCW Workspaces                       [+] │
├──────────────────────────────────────────┤
│ ● ccw/menubarapp          5m ago     [>] │  ← connected (has client)
│ ◐ myapp/feature-x         1h ago     [>] │  ← alive (no client)
│ ○ myapp/fix-auth          2h ago     [>] │  ← dead
├──────────────────────────────────────────┤
│ Stale Workspaces (2)                 [>] │
├──────────────────────────────────────────┤
│ Settings...                         ⌘,  │
│ Quit                                ⌘Q  │
└──────────────────────────────────────────┘
```
**Status indicators:**
- 🟢 Connected (has clients) → Click = focus existing iTerm window
- 🟡 Alive (no clients) → Click = attach in new maximized iTerm
- 🔴 Dead → Click = recreate session + attach

[>] hover shows: Open, Open (no resume), Info, Remove

### New Workspace Window
```
┌────────────────────────────────────┐
│       Create New Workspace         │
├────────────────────────────────────┤
│ Repository:  [ccw            ▼]    │
│ Branch:      [feature/...     ]    │
│ Base branch: [main           ▼]    │
│ [ ] Open immediately               │
│                                    │
│        [Cancel]    [Create]        │
└────────────────────────────────────┘
```

### Settings Window
```
┌────────────────────────────────────┐
│          CCW Settings              │
├────────────────────────────────────┤
│ Repos Directory: [~/github    ][…] │
│                                    │
│ Layout:                            │
│   Left:  [claude  ▼]               │
│   Right: [lazygit ▼]               │
│                                    │
│ [x] iTerm CC Mode                  │
│ [ ] Skip permission prompts        │
│                                    │
│   [Reset Defaults]     [Save]      │
└────────────────────────────────────┘
```

---

## Implementation Sequence

### Phase 1: Core Infrastructure
1. Create Xcode project (macOS 13+, SwiftUI lifecycle)
2. Implement Models: `Workspace`, `WorkspaceStatus`, `CCWConfig`
3. Build `CLIBridge` with Process execution
4. Create `AppState` with workspace listing
5. Basic menubar with `MenuBarExtra`

### Phase 2: Menubar UI
1. `MenuBarView` with workspace list
2. `WorkspaceRow` with status indicator
3. `StatusIndicator` component (green/red dot)
4. Open action → `ccw open`
5. Refresh on menu open (`.onAppear` trigger)

### Phase 3: Workspace Management
1. `NewWorkspaceWindow` form UI
2. Repo discovery (list dirs in repos_dir)
3. Remove with confirmation dialog
4. Workspace info popover

### Phase 4: Settings & Config
1. `SettingsWindow` with all config options
2. Path picker for repos_dir
3. Layout selector dropdown
4. Config persistence via CLI

### Phase 5: Onboarding
1. Detect ccw installation
2. Check `config.onboarded` flag
3. `OnboardingWindow` wizard
4. Dependency checker UI

---

## Onboarding Flow

### Detection Logic (App Launch)
```swift
// In AppState.init()
func checkSetupStatus() async {
    // 1. Check if bundled ccw exists
    guard let ccwPath = Bundle.main.path(forResource: "ccw", ofType: nil) else {
        state = .missingCLI
        return
    }

    // 2. Check dependencies
    let deps = try await cli.checkDependencies()  // git, tmux, claude, iterm
    if !deps.allInstalled {
        state = .missingDependencies(deps)
        return
    }

    // 3. Check if onboarded
    if let config = try? await cli.getConfig(), config.onboarded {
        state = .ready
    } else {
        state = .needsOnboarding
    }
}
```

### Onboarding Window UI
```
┌─────────────────────────────────────────────┐
│           Welcome to CCW Menubar            │
├─────────────────────────────────────────────┤
│                                             │
│  Let's set up your workspace environment.   │
│                                             │
│  Dependencies:                              │
│  ✓ git       installed                      │
│  ✓ tmux      installed                      │
│  ✓ iTerm2    installed                      │
│  ✓ claude    installed                      │
│  ○ lazygit   optional (not found)           │
│                                             │
│  ─────────────────────────────────────────  │
│                                             │
│  Where are your git repositories?           │
│  [~/github                           ][…]   │
│                                             │
│  Pane layout:                               │
│  ( ) claude | lazygit  (default)            │
│  ( ) lazygit | claude                       │
│                                             │
│  [ ] Skip Claude permission prompts         │
│      (auto-accept all tool use)             │
│                                             │
│                     [Complete Setup]        │
└─────────────────────────────────────────────┘
```

### Onboarding States

| State | Menubar Behavior |
|-------|------------------|
| `missingCLI` | Show error icon, click opens alert with instructions |
| `missingDependencies` | Show warning icon, click opens dependency checker |
| `needsOnboarding` | Show setup icon, click opens OnboardingWindow |
| `ready` | Normal operation |

### After Onboarding Completes
1. Save config: `ccw config repos_dir <path> && ccw config onboarded true`
2. Refresh AppState
3. Show main workspace list
4. Optional: Show "Getting Started" tooltip

### Re-running Onboarding
- Settings → "Re-run Setup" button
- Sets `onboarded = false` and reopens OnboardingWindow

---

### Phase 6: Polish
1. Keyboard shortcuts (KeyboardShortcuts package)
2. Stale workspace detection/cleanup
3. Launch at login option
4. Error handling and alerts

---

## Dependencies

**System Dependencies (Required):**
- iTerm2
- tmux
- git
- claude

**System Dependencies (Optional):**
- lazygit

**Required (Built-in):**
- SwiftUI, Foundation, AppKit, Combine

**Optional (SPM):**
- `sindresorhus/KeyboardShortcuts` - Global hotkeys
- `sindresorhus/LaunchAtLogin` - Login item

**Build Requirements:**
- macOS 13.0+ (Ventura)
- Xcode 15.0+
- Swift 5.9+

---

## Sandboxing & Permissions (Decision: No App Sandbox)

- **App Sandbox**: Disabled. Use hardened runtime + Developer ID signing for distribution.
- **Runtime PATH**: Explicitly set `PATH` in `CLIBridge` so bundled `ccw` can find `git`, `tmux`, `claude`, and `lazygit` (e.g., include `/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin`).
- **Automation**: Add `NSAppleEventsUsageDescription` to Info.plist and sign with the `com.apple.security.automation.apple-events` entitlement so AppleScript can control iTerm.
- **First-run prompt**: Expect the macOS Automation permission prompt the first time the app focuses iTerm.

---

## Distribution & Installation

### Current CLI Distribution
The CLI is distributed via:
- **GitHub Releases**: Binary downloads for darwin/linux (amd64/arm64)
- **Homebrew Tap**: `brew install justanotheratom/tap/ccw`

### Menubar App Distribution Options

**Bundled CLI in App**
```
CCWMenubar.app/
├── Contents/
│   ├── MacOS/
│   │   ├── CCWMenubar        # Swift app binary
│   │   └── ccw               # Embedded Go CLI binary
│   └── Resources/
```

- App bundle includes the `ccw` binary
- CLIBridge uses ONLY bundled binary: `Bundle.main.url(forAuxiliaryExecutable: "ccw")`
- No system fallback - ensures version consistency
- Single download, CLI is fully self-contained

### Installation Methods

**Primary: Homebrew Cask**
```bash
brew install --cask justanotheratom/tap/ccw-menubar
```

**Cask Formula** (`packaging/homebrew/ccw-menubar.rb`):
```ruby
cask "ccw-menubar" do
  version "1.0.0"
  sha256 "..."

  url "https://github.com/justanotheratom/ccw/releases/download/v#{version}/CCWMenubar.dmg"
  name "CCW Menubar"
  desc "Menubar app for Claude Code Workspace manager"
  homepage "https://github.com/justanotheratom/ccw"

  app "CCWMenubar.app"

  zap trash: [
    "~/.ccw",  # Config and registry (optional cleanup)
  ]
end
```

**Alternative: Direct DMG Download**
- Download from GitHub Releases
- Drag CCWMenubar.app to /Applications

### Build & Release Workflow (GitHub Actions)

**New file: `.github/workflows/release-menubar.yml`**
```yaml
name: Release Menubar App

on:
  push:
    tags:
      - "menubar-v*"

jobs:
  build:
    runs-on: macos-14  # Apple Silicon runner
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Build ccw CLI (universal binary)
        run: |
          GOOS=darwin GOARCH=arm64 go build -o ccw-arm64 .
          GOOS=darwin GOARCH=amd64 go build -o ccw-amd64 .
          lipo -create -output ccw ccw-arm64 ccw-amd64

      - name: Set up Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: '15.0'

      - name: Build Menubar App
        run: |
          cd CCWMenubar
          xcodebuild -project CCWMenubar.xcodeproj \
            -scheme CCWMenubar \
            -configuration Release \
            -archivePath build/CCWMenubar.xcarchive \
            archive
          xcodebuild -exportArchive \
            -archivePath build/CCWMenubar.xcarchive \
            -exportPath build/export \
            -exportOptionsPlist ExportOptions.plist

      - name: Embed CLI in App Bundle
        run: |
          cp ccw CCWMenubar/build/export/CCWMenubar.app/Contents/MacOS/

      - name: Code Sign (if certificates available)
        if: env.APPLE_CERTIFICATE != ''
        run: |
          # Import certificate, sign embedded CLI, then sign app with hardened runtime
          codesign --force --sign "$DEVELOPER_ID" \
            CCWMenubar/build/export/CCWMenubar.app/Contents/MacOS/ccw
          codesign --force --options runtime \
            --entitlements CCWMenubar/Entitlements.plist \
            --sign "$DEVELOPER_ID" \
            CCWMenubar/build/export/CCWMenubar.app

      - name: Create DMG
        run: |
          brew install create-dmg
          create-dmg \
            --volname "CCW Menubar" \
            --window-pos 200 120 \
            --window-size 600 400 \
            --icon-size 100 \
            --app-drop-link 425 178 \
            CCWMenubar.dmg \
            CCWMenubar/build/export/CCWMenubar.app

      - name: Notarize and Staple (if certificates available)
        if: env.APPLE_CERTIFICATE != ''
        run: |
          xcrun notarytool submit CCWMenubar.dmg --wait \
            --key "$NOTARYTOOL_KEY" --key-id "$NOTARYTOOL_KEY_ID" --issuer "$NOTARYTOOL_ISSUER"
          xcrun stapler staple CCWMenubar.dmg

      - name: Upload to Release
        uses: softprops/action-gh-release@v2
        with:
          files: CCWMenubar.dmg
          generate_release_notes: true

      - name: Update Homebrew Tap
        run: |
          # Trigger tap update workflow
          gh workflow run update-tap.yml -f app=menubar

env:
  APPLE_CERTIFICATE: ${{ secrets.APPLE_CERTIFICATE }}
  DEVELOPER_ID: ${{ secrets.DEVELOPER_ID }}
  NOTARYTOOL_KEY: ${{ secrets.NOTARYTOOL_KEY }}
  NOTARYTOOL_KEY_ID: ${{ secrets.NOTARYTOOL_KEY_ID }}
  NOTARYTOOL_ISSUER: ${{ secrets.NOTARYTOOL_ISSUER }}
```

### Code Signing Requirements
- **Signed builds**: Apple Developer ID certificate ($99/year) with hardened runtime, notarized for Gatekeeper
- **Unsigned builds**: Users run `xattr -cr /Applications/CCWMenubar.app` to bypass Gatekeeper
- **Entitlements**: Include `com.apple.security.automation.apple-events` for iTerm control
- **Info.plist**: Add `NSAppleEventsUsageDescription` (Automation prompt)

---

## CLI Enhancements Required

These changes to the Go CLI are required before the menubar app can work:

### 1. Enhanced JSON Output for `ccw ls`

**File:** `cmd/ls.go`, `internal/workspace/manager.go`

Add `HasClients` field to detect connected vs alive:
```go
// internal/workspace/manager.go
type WorkspaceStatus struct {
    ID           string
    Workspace    Workspace
    SessionAlive bool
    HasClients   bool  // NEW: tmux session has attached clients
}

func (m *Manager) ListWorkspaces(ctx context.Context) ([]WorkspaceStatus, error) {
    // ... existing code ...
    for i, st := range statuses {
        if st.SessionAlive {
            hasClients, _ := m.tmux.HasAttachedClients(st.Workspace.TmuxSession)
            statuses[i].HasClients = hasClients
        }
    }
    return statuses, nil
}
```

### 2. Client Detection in Tmux

**File:** `internal/tmux/tmux.go`

```go
func (r Runner) HasAttachedClients(session string) (bool, error) {
    out, err := r.run(context.Background(), "list-clients", "-t", session)
    if err != nil {
        return false, nil
    }
    return strings.TrimSpace(out) != "", nil
}
```

### 3. Smart Session Attachment

**File:** `internal/tmux/tmux.go`

Update `AttachSession` to focus existing window if connected:
```go
func (r Runner) AttachSession(name string) error {
    hasClients, _ := r.HasAttachedClients(name)
    if hasClients {
        return focusExistingMacWindow(name)  // NEW
    }
    return openNewMacTerminalWindow(name, r.PreferCC)
}

func focusExistingMacWindow(session string) error {
    script := fmt.Sprintf(`tell application "iTerm"
        repeat with w in windows
            if name of w contains "%s" then
                set frontmost of w to true
                activate
                return
            end if
        end repeat
    end tell`, session)
    return runOsaScript(script)
}
```

Note: Ensure iTerm window identification is reliable (window title contains the session name or query iTerm sessions instead of window titles).

### 4. New Commands

**File:** `cmd/stale.go`
- Add `--json` flag for JSON output

**File:** `cmd/info.go`
- Add `--json` flag for JSON output

**File:** `cmd/repos.go` (NEW)
- `ccw repos` - List directories in repos_dir for "New Workspace" dropdown

### 5. Dependency Check Command

**File:** `cmd/check.go` (NEW)
- `ccw check --json` - Returns dependency status for onboarding
```json
{
  "git": {"installed": true, "path": "/usr/bin/git"},
  "tmux": {"installed": true, "path": "/opt/homebrew/bin/tmux"},
  "iterm": {"installed": true, "path": "/Applications/iTerm.app"},
  "claude": {"installed": true, "path": "/usr/local/bin/claude"},
  "lazygit": {"installed": false, "path": "", "optional": true}
}
```

### 6. Non-interactive Flags for GUI Use

**Files:** `cmd/open.go`, `cmd/new.go`, `cmd/rm.go`
- Ensure the app can invoke commands without prompts (`open --no-resume`, `new --no-attach`, `rm --yes`)

---

## Implementation Tasks

### Task Dependency Graph

```
CLI-1 ──┬──> CLI-2 ──> CLI-3 ──> CLI-4
        │                         │
        ├──> CLI-5                │
        ├──> CLI-6                │
        └──> CLI-7                │
                                  v
APP-0 ──> APP-1 ──> APP-2 ──> APP-3 ──> APP-4 ──> APP-5 ──> APP-6
                                  │
                                  v
BUILD-1 ──> BUILD-2 ──> BUILD-3 ──> RELEASE-1
```

**Legend:**
- `CLI-*`: Go CLI enhancements (required first)
- `APP-*`: Swift menubar app development
- `BUILD-*`: Build infrastructure
- `RELEASE-*`: Release automation

---

### Phase 1: CLI Enhancements

#### CLI-1: Add HasClients detection to tmux package
**File:** `internal/tmux/tmux.go`
**Depends on:** None
**Description:** Add method to detect if a tmux session has attached clients.

```go
// Add this method to Runner struct
func (r Runner) HasAttachedClients(session string) (bool, error) {
    out, err := r.run(context.Background(), "list-clients", "-t", session)
    if err != nil {
        return false, nil  // Session doesn't exist or error
    }
    return strings.TrimSpace(out) != "", nil
}
```

**Test:** Run `ccw ls`, then manually run `tmux list-clients -t <session>` to verify.

---

#### CLI-2: Add HasClients to WorkspaceStatus
**File:** `internal/workspace/manager.go`
**Depends on:** CLI-1
**Description:** Extend WorkspaceStatus struct and populate HasClients in ListWorkspaces.

1. Add field to struct:
```go
type WorkspaceStatus struct {
    ID           string    `json:"ID"`
    Workspace    Workspace `json:"Workspace"`
    SessionAlive bool      `json:"SessionAlive"`
    HasClients   bool      `json:"HasClients"`  // ADD THIS
}
```

2. Populate in ListWorkspaces:
```go
func (m *Manager) ListWorkspaces(ctx context.Context) ([]WorkspaceStatus, error) {
    // ... existing code to load workspaces ...

    for i, st := range statuses {
        if st.SessionAlive {
            hasClients, _ := m.tmux.HasAttachedClients(st.Workspace.TmuxSession)
            statuses[i].HasClients = hasClients
        }
    }
    return statuses, nil
}
```

**Test:** Run `ccw ls --json` and verify `HasClients` field appears.

---

#### CLI-3: Smart session attachment
**File:** `internal/tmux/tmux.go`
**Depends on:** CLI-1
**Description:** Update AttachSession to focus existing window if connected.

1. Add focusExistingMacWindow function:
```go
func focusExistingMacWindow(session string) error {
    script := fmt.Sprintf(`tell application "iTerm"
        repeat with w in windows
            if name of w contains "%s" then
                set frontmost of w to true
                activate
                return
            end if
        end repeat
    end tell`, session)
    return runOsaScript(script)
}
```

2. Update AttachSession:
```go
func (r Runner) AttachSession(name string) error {
    hasClients, _ := r.HasAttachedClients(name)
    if hasClients {
        return focusExistingMacWindow(name)
    }
    // Existing code for creating new window...
    return openNewMacTerminalWindow(name, r.PreferCC)
}
```

**Test:** Open a workspace, then run `ccw open <workspace>` again - should focus existing window.

---

#### CLI-4: Add --json flags to stale and info commands
**Files:** `cmd/stale.go`, `cmd/info.go`
**Depends on:** None (parallel with CLI-1,2,3)
**Description:** Add JSON output option to existing commands.

For `cmd/stale.go`:
```go
func init() {
    rootCmd.AddCommand(staleCmd)
    staleCmd.Flags().Bool("json", false, "Output as JSON")  // ADD
    staleCmd.Flags().Bool("rm", false, "Remove stale workspaces")
    staleCmd.Flags().Bool("force", false, "Force removal")
}

// In RunE, add:
showJSON, _ := cmd.Flags().GetBool("json")
if showJSON {
    enc := json.NewEncoder(cmd.OutOrStdout())
    enc.SetIndent("", "  ")
    return enc.Encode(stale)
}
```

For `cmd/info.go`:
```go
infoCmd.Flags().Bool("json", false, "Output as JSON")

showJSON, _ := cmd.Flags().GetBool("json")
if showJSON {
    enc := json.NewEncoder(cmd.OutOrStdout())
    enc.SetIndent("", "  ")
    return enc.Encode(status)
}
// Existing human-readable output...
```

**Test:** Run `ccw stale --json` and `ccw info <workspace> --json`.

**Info JSON example:**
```json
{
  "ID": "repo/branch",
  "Workspace": {
    "repo": "repo",
    "repo_path": "/Users/me/dev/repo",
    "branch": "branch",
    "base_branch": "main",
    "worktree_path": "/Users/me/dev/repo-branch",
    "claude_session": "ccw-repo-branch",
    "tmux_session": "ccw-repo-branch",
    "created_at": "2024-01-01T12:00:00Z",
    "last_accessed_at": "2024-01-10T09:30:00Z"
  },
  "SessionAlive": true,
  "HasClients": false
}
```

---

#### CLI-5: Add new commands (repos, check)
**Files:** `cmd/repos.go` (new), `cmd/check.go` (new)
**Depends on:** None (parallel with CLI-1,2,3)
**Description:** Add repos listing and dependency check (including iTerm detection); `check --json` emits stable JSON for onboarding.

**cmd/repos.go:**
```go
package cmd

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
    Use:   "repos",
    Short: "List available repositories",
    RunE: func(cmd *cobra.Command, args []string) error {
        mgr, err := newManager()
        if err != nil {
            return err
        }
        cfg := mgr.GetConfig()
        reposDir := cfg.ExpandedReposDir()

        entries, err := os.ReadDir(reposDir)
        if err != nil {
            return err
        }

        var repos []string
        for _, e := range entries {
            if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
                repos = append(repos, e.Name())
            }
        }

        showJSON, _ := cmd.Flags().GetBool("json")
        if showJSON {
            return json.NewEncoder(cmd.OutOrStdout()).Encode(repos)
        }
        for _, r := range repos {
            fmt.Println(r)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(reposCmd)
    reposCmd.Flags().Bool("json", false, "Output as JSON")
}
```

**cmd/check.go:**
```go
package cmd

import (
    "encoding/json"
    "fmt"
    "github.com/ccw/ccw/internal/deps"
    "github.com/spf13/cobra"
)

type DepStatus struct {
    Installed bool   `json:"installed"`
    Path      string `json:"path"`
    Optional  bool   `json:"optional,omitempty"`
}

var checkCmd = &cobra.Command{
    Use:   "check",
    Short: "Check dependencies",
    RunE: func(cmd *cobra.Command, args []string) error {
        result := make(map[string]DepStatus)
        for _, d := range deps.All() {
            path, err := d.Check()
            result[d.Name] = DepStatus{
                Installed: err == nil,
                Path:      path,
                Optional:  d.Name == "lazygit",
            }
        }
        showJSON, _ := cmd.Flags().GetBool("json")
        if showJSON {
            return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
        }
        for name, st := range result {
            fmt.Fprintf(cmd.OutOrStdout(), "%s\t%t\t%s\n", name, st.Installed, st.Path)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(checkCmd)
    checkCmd.Flags().Bool("json", false, "Output as JSON")
}
```

**iTerm detection (deps):**
- Add `iterm` to `deps.DefaultDependencies()` with `DisplayName: "iTerm2"` and install hint.
- Special-case app bundle detection in `deps.Check`:
```go
func Check(dep Dependency) Result {
    if dep.Name == "iterm" {
        if path := findITermApp(); path != "" {
            return Result{Dependency: dep, Found: true, Path: path}
        }
        return Result{Dependency: dep, Found: false}
    }
    path, err := exec.LookPath(dep.Name)
    if err != nil {
        return Result{Dependency: dep, Found: false}
    }
    return Result{Dependency: dep, Found: true, Path: path}
}

func findITermApp() string {
    candidates := []string{
        "/Applications/iTerm.app",
        "/Applications/iTerm2.app",
        filepath.Join(os.Getenv("HOME"), "Applications/iTerm.app"),
        filepath.Join(os.Getenv("HOME"), "Applications/iTerm2.app"),
    }
    for _, path := range candidates {
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }
    return ""
}
```

**Test:** Run `ccw repos --json` and `ccw check --json`.

---

#### CLI-6: Audit non-interactive flags used by the app
**Files:** `cmd/open.go`, `cmd/new.go`, `cmd/rm.go` (as needed)
**Depends on:** None (parallel with CLI-1,2,3)
**Description:** Ensure GUI-triggered commands never prompt for input (e.g., add/verify `open --no-resume`, `new --no-attach`, `rm --yes`).

**Test:** Run each command with stdin closed (e.g., `</dev/null`) and confirm no prompts.

---

#### CLI-7: Make iTerm window focus reliable
**Files:** `internal/tmux/tmux.go` (and any iTerm helpers)
**Depends on:** CLI-3
**Description:** Ensure the focused iTerm window can be uniquely identified (e.g., set window/title to include tmux session name or query iTerm sessions instead of window titles).

**Test:** Open a workspace twice and confirm the existing iTerm window is focused every time.

---

### Phase 2: Swift Menubar App

#### APP-0: CLI parity checklist
**Depends on:** CLI-1 through CLI-7
**Description:** Map CLI commands/flags to menubar actions, confirm required JSON schemas, and mark any CLI features that will remain CLI-only.

---

#### APP-1: Create Xcode project and data models
**Depends on:** APP-0, CLI-2, CLI-4, CLI-5 (final JSON schema for ls/stale/info/check)
**Description:** Set up Swift project with Codable models matching CLI output (ls, stale, info, config, check).

1. Create new Xcode project:
   - File > New > Project > macOS > App
   - Product Name: CCWMenubar
   - Interface: SwiftUI
   - Life Cycle: SwiftUI App

2. Create `Models/Workspace.swift`:
```swift
import Foundation

struct Workspace: Codable {
    let repo: String
    let repoPath: String
    let branch: String
    let baseBranch: String
    let worktreePath: String
    let claudeSession: String
    let tmuxSession: String
    let createdAt: Date
    let lastAccessedAt: Date

    enum CodingKeys: String, CodingKey {
        case repo
        case repoPath = "repo_path"
        case branch
        case baseBranch = "base_branch"
        case worktreePath = "worktree_path"
        case claudeSession = "claude_session"
        case tmuxSession = "tmux_session"
        case createdAt = "created_at"
        case lastAccessedAt = "last_accessed_at"
    }
}

struct WorkspaceStatus: Codable, Identifiable {
    let id: String
    let workspace: Workspace
    let sessionAlive: Bool
    let hasClients: Bool

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case workspace = "Workspace"
        case sessionAlive = "SessionAlive"
        case hasClients = "HasClients"
    }

    var state: WorkspaceState {
        if !sessionAlive { return .dead }
        if hasClients { return .connected }
        return .alive
    }
}

enum WorkspaceState {
    case connected  // Has attached clients
    case alive      // Session exists, no clients
    case dead       // No session
}
```

3. Create `Models/CCWConfig.swift`:
```swift
struct CCWConfig: Codable {
    var version: Int
    var reposDir: String
    var itermCCMode: Bool
    var claudeRenameDelay: Int
    var layout: Layout
    var onboarded: Bool
    var claudeDangerouslySkipPermissions: Bool

    struct Layout: Codable {
        var left: String
        var right: String
    }

    enum CodingKeys: String, CodingKey {
        case version
        case reposDir = "repos_dir"
        case itermCCMode = "iterm_cc_mode"
        case claudeRenameDelay = "claude_rename_delay"
        case layout
        case onboarded
        case claudeDangerouslySkipPermissions = "claude_dangerously_skip_permissions"
    }
}
```

4. Create `Models/WorkspaceInfo.swift` and `Models/DepStatus.swift` to match `ccw info --json` and `ccw check --json`.

**Test:** Write unit tests that decode sample JSON from `ccw ls --json`, `ccw stale --json`, and `ccw info --json`.

---

#### APP-2: Implement CLIBridge
**Depends on:** APP-1, CLI-5, CLI-6
**Description:** Create actor to execute ccw commands and parse output.

Create `Core/CLI/CLIBridge.swift`:
```swift
import Foundation

actor CLIBridge {
    enum CLIError: Error, LocalizedError {
        case commandFailed(String)
        case parseError(String)
        case ccwNotFound

        var errorDescription: String? {
            switch self {
            case .commandFailed(let msg): return "Command failed: \(msg)"
            case .parseError(let msg): return "Parse error: \(msg)"
            case .ccwNotFound: return "ccw binary not found in app bundle"
            }
        }
    }

    private let ccwURL: URL
    private let decoder: JSONDecoder

    init() throws {
        guard let url = Bundle.main.url(forAuxiliaryExecutable: "ccw") else {
            throw CLIError.ccwNotFound
        }
        self.ccwURL = url
        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601
    }

    func listWorkspaces() async throws -> [WorkspaceStatus] {
        let output = try await execute(["ls", "--json"])
        return try decoder.decode([WorkspaceStatus].self, from: output)
    }

    func openWorkspace(_ id: String, resume: Bool = true) async throws {
        var args = ["open", id]
        if !resume { args.append("--no-resume") }
        _ = try await execute(args)
    }

    func createWorkspace(repo: String, branch: String, base: String?) async throws {
        var args = ["new", repo, branch, "--no-attach"]
        if let base = base {
            args.append(contentsOf: ["--base", base])
        }
        _ = try await execute(args)
    }

    func removeWorkspace(_ id: String, force: Bool) async throws {
        var args = ["rm", id, "--yes"]
        if force { args.append("--force") }
        _ = try await execute(args)
    }

    func staleWorkspaces() async throws -> [WorkspaceStatus] {
        let output = try await execute(["stale", "--json"])
        return try decoder.decode([WorkspaceStatus].self, from: output)
    }

    func workspaceInfo(_ id: String) async throws -> WorkspaceInfo {
        let output = try await execute(["info", id, "--json"])
        return try decoder.decode(WorkspaceInfo.self, from: output)
    }

    func listRepos() async throws -> [String] {
        let output = try await execute(["repos", "--json"])
        return try decoder.decode([String].self, from: output)
    }

    func checkDependencies() async throws -> [String: DepStatus] {
        let output = try await execute(["check", "--json"])
        return try decoder.decode([String: DepStatus].self, from: output)
    }

    func getConfig() async throws -> CCWConfig {
        let configPath = FileManager.default.homeDirectoryForCurrentUser
            .appendingPathComponent(".ccw/config.json")
        let data = try Data(contentsOf: configPath)
        return try decoder.decode(CCWConfig.self, from: data)
    }

    func setConfig(key: String, value: String) async throws {
        _ = try await execute(["config", key, value])
    }

    private func execute(_ arguments: [String]) async throws -> Data {
        let process = Process()
        process.executableURL = ccwURL
        process.arguments = arguments
        var env = ProcessInfo.processInfo.environment
        env["PATH"] = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
        process.environment = env

        let stdout = Pipe()
        let stderr = Pipe()
        process.standardOutput = stdout
        process.standardError = stderr

        try process.run()
        process.waitUntilExit()

        if process.terminationStatus != 0 {
            let errorData = stderr.fileHandleForReading.readDataToEndOfFile()
            let errorMsg = String(data: errorData, encoding: .utf8) ?? "Unknown error"
            throw CLIError.commandFailed(errorMsg)
        }

        return stdout.fileHandleForReading.readDataToEndOfFile()
    }
}

struct DepStatus: Codable {
    let installed: Bool
    let path: String
    let optional: Bool?
}
```

**Test:** Run in Xcode with placeholder ccw binary, verify commands execute.

---

#### APP-3: Create AppState and basic MenuBarView
**Depends on:** APP-2
**Description:** Set up centralized state and menubar UI. Use a menu-open hook (`MenuBarExtra(isPresented:)` or an `NSMenuDelegate`) so refresh triggers every time the menu is opened (not just first `onAppear`). Update the status item icon based on the highest-priority workspace state.

Create `ViewModels/AppState.swift`:
```swift
import SwiftUI

@MainActor
class AppState: ObservableObject {
    @Published var workspaces: [WorkspaceStatus] = []
    @Published var isLoading = false
    @Published var error: Error?
    @Published var setupState: SetupState = .checking

    enum SetupState {
        case checking
        case needsOnboarding
        case missingDependencies([String: DepStatus])
        case ready
        case error(String)
    }

    private var cli: CLIBridge?

    init() {
        Task { await initialize() }
    }

    private func initialize() async {
        do {
            cli = try CLIBridge()
            let deps = try await cli!.checkDependencies()
            if deps.contains(where: { !$0.value.installed && $0.value.optional != true }) {
                setupState = .missingDependencies(deps)
                return
            }
            let config = try await cli!.getConfig()
            if !config.onboarded {
                setupState = .needsOnboarding
            } else {
                setupState = .ready
                await refreshWorkspaces()
            }
        } catch CLIBridge.CLIError.ccwNotFound {
            setupState = .error("ccw binary not found in app bundle")
        } catch {
            setupState = .needsOnboarding
        }
    }

    func refreshWorkspaces() async {
        guard let cli = cli else { return }
        isLoading = true
        defer { isLoading = false }

        do {
            workspaces = try await cli.listWorkspaces()
            error = nil
        } catch {
            self.error = error
        }
    }

    func openWorkspace(_ id: String) async {
        guard let cli = cli else { return }
        do {
            try await cli.openWorkspace(id)
        } catch {
            self.error = error
        }
    }
}
```

Update `CCWMenubarApp.swift`:
```swift
import SwiftUI

@main
struct CCWMenubarApp: App {
    @StateObject private var appState = AppState()

    var body: some Scene {
        MenuBarExtra("CCW", systemImage: "terminal.fill") {
            MenuBarView()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView()
                .environmentObject(appState)
        }
    }
}
```

Create `Views/MenuBar/MenuBarView.swift`:
```swift
import SwiftUI

struct MenuBarView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Header
            HStack {
                Text("CCW Workspaces")
                    .font(.headline)
                Spacer()
                Button(action: { /* show new workspace */ }) {
                    Image(systemName: "plus")
                }
            }
            .padding()

            Divider()

            // Workspace list
            if appState.isLoading {
                ProgressView()
                    .padding()
            } else {
                ForEach(appState.workspaces) { workspace in
                    WorkspaceRow(workspace: workspace)
                }
            }

            Divider()

            // Footer
            Button("Settings...") {
                NSApp.sendAction(Selector(("showSettingsWindow:")), to: nil, from: nil)
            }
            .keyboardShortcut(",", modifiers: .command)

            Button("Quit") {
                NSApp.terminate(nil)
            }
            .keyboardShortcut("q", modifiers: .command)
        }
        .frame(width: 350)
        .onAppear {
            Task { await appState.refreshWorkspaces() }
        }
    }
}
```

Create `Views/MenuBar/WorkspaceRow.swift`:
```swift
import SwiftUI

struct WorkspaceRow: View {
    let workspace: WorkspaceStatus
    @EnvironmentObject var appState: AppState

    var body: some View {
        Button(action: {
            Task { await appState.openWorkspace(workspace.id) }
        }) {
            HStack {
                statusIndicator
                Text(workspace.id)
                Spacer()
                Text(timeAgo)
                    .foregroundColor(.secondary)
                    .font(.caption)
            }
            .padding(.horizontal)
            .padding(.vertical, 6)
        }
        .buttonStyle(.plain)
    }

    private var statusIndicator: some View {
        Circle()
            .fill(statusColor)
            .frame(width: 8, height: 8)
    }

    private var statusColor: Color {
        switch workspace.state {
        case .connected: return .green
        case .alive: return .yellow
        case .dead: return .red
        }
    }

    private var timeAgo: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: workspace.workspace.lastAccessedAt, relativeTo: Date())
    }
}
```

Add a context menu for each workspace row: Open, Open (no resume), Info, Remove. Wire `Open (no resume)` to `openWorkspace(resume: false)`, `Info` to `workspaceInfo`, and refresh after destructive actions.

**Test:** Build and run, verify menubar icon appears and shows workspaces.

---

#### APP-4: Implement New Workspace and Remove workflows
**Depends on:** APP-3
**Description:** Add sheets for creating and removing workspaces; refresh the workspace list after create/remove and surface errors.

Create `Views/Windows/NewWorkspaceView.swift`:
```swift
import SwiftUI

struct NewWorkspaceView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var repos: [String] = []
    @State private var selectedRepo = ""
    @State private var branchName = ""
    @State private var baseBranch = "main"
    @State private var isCreating = false

    var body: some View {
        Form {
            Picker("Repository", selection: $selectedRepo) {
                ForEach(repos, id: \.self) { repo in
                    Text(repo).tag(repo)
                }
            }

            TextField("Branch name", text: $branchName)

            TextField("Base branch", text: $baseBranch)

            HStack {
                Button("Cancel") { dismiss() }
                Spacer()
                Button("Create") { create() }
                    .disabled(selectedRepo.isEmpty || branchName.isEmpty)
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding()
        .frame(width: 400)
        .task { await loadRepos() }
    }

    private func loadRepos() async {
        // Load from CLI
    }

    private func create() {
        isCreating = true
        Task {
            // await appState.createWorkspace(...)
            dismiss()
        }
    }
}
```

**Test:** Open New Workspace sheet, create a workspace, verify it appears in list.

---

#### APP-5: Implement Settings and Onboarding
**Depends on:** APP-3
**Description:** Settings window and first-run onboarding flow. Include iTerm in dependency checks and show Automation permission guidance if AppleScript fails.

Create `Views/Windows/SettingsView.swift` and `Views/Windows/OnboardingView.swift` following the UI mockups in the plan.

Key points:
- Settings reads/writes via `appState.cli.setConfig()`
- Onboarding checks dependencies via `ccw check`
- After onboarding, call `ccw config onboarded true`

**Test:** Delete `~/.ccw/config.json`, relaunch app, verify onboarding appears.

---

#### APP-6: Polish (keyboard shortcuts, stale workspaces)
**Depends on:** APP-5
**Description:** Add remaining features.

1. Add SPM dependency: `sindresorhus/KeyboardShortcuts`
2. Implement global hotkey to show/hide menu
3. Add Stale Workspaces submenu
4. Add Launch at Login option via `sindresorhus/LaunchAtLogin`

---

### Phase 3: Build & Release

#### BUILD-1: Set up Xcode project structure
**Depends on:** APP-1
**Description:** Configure Xcode project for embedding ccw binary and distribution settings.

1. Add "Run Script" build phase to copy ccw into app bundle (Debug + Release)
2. Add `Entitlements.plist` with `com.apple.security.automation.apple-events`; disable App Sandbox; enable Hardened Runtime (Release)
3. Update Info.plist with `LSUIElement` and `NSAppleEventsUsageDescription`
4. Create `ExportOptions.plist` for archiving
5. Add app icon assets

---

#### BUILD-2: Create GitHub Actions workflow
**Depends on:** BUILD-1, CLI-1 through CLI-7
**Description:** Automate the build process, including signing and notarization.

Create `.github/workflows/release-menubar.yml` as shown in the plan.

---

#### BUILD-3: Create Homebrew cask formula
**Depends on:** BUILD-2
**Description:** Add cask for installation.

Create `packaging/homebrew/ccw-menubar.rb` as shown in the plan.

---

#### RELEASE-1: Test and release
**Depends on:** BUILD-3
**Description:** Create first release.

1. Tag with `menubar-v1.0.0`
2. Verify GitHub Actions builds DMG
3. Test installation via `brew install --cask`
4. Update README with installation instructions

---

## Critical Reference Files

| File | Purpose |
|------|---------|
| `cmd/ls.go` | JSON output format for workspace list |
| `internal/workspace/registry.go` | Workspace model structure |
| `internal/config/config.go` | Config fields and defaults |
| `internal/onboarding/onboarding.go` | Onboarding flow logic |
