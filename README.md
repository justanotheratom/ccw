# CCWMenubar - macOS App

A modern macOS application using a **workspace + SPM package** architecture for clean separation between app shell and feature code.

## Project Architecture

```
CCWMenubar/
├── CCWMenubar.xcworkspace/              # Open this file in Xcode
├── CCWMenubar.xcodeproj/                # App shell project
├── CCWMenubar/                          # App target (minimal)
│   ├── Assets.xcassets/                # App-level assets (icons, colors)
│   ├── CCWMenubarApp.swift              # App entry point
│   ├── CCWMenubar.entitlements          # App sandbox settings
│   └── CCWMenubar.xctestplan            # Test configuration
├── CCWMenubarPackage/                   # 🚀 Primary development area
│   ├── Package.swift                   # Package configuration
│   ├── Sources/CCWMenubarFeature/       # Your feature code
│   └── Tests/CCWMenubarFeatureTests/    # Unit tests
└── CCWMenubarUITests/                   # UI automation tests
```

## Key Architecture Points

### Workspace + SPM Structure
- **App Shell**: `CCWMenubar/` contains minimal app lifecycle code
- **Feature Code**: `CCWMenubarPackage/Sources/CCWMenubarFeature/` is where most development happens
- **Separation**: Business logic lives in the SPM package, app target just imports and displays it

### Buildable Folders (Xcode 16)
- Files added to the filesystem automatically appear in Xcode
- No need to manually add files to project targets
- Reduces project file conflicts in teams

### App Sandbox
The menubar app runs without App Sandbox and uses the hardened runtime for distribution. Apple Events automation is enabled for iTerm control in `Config/CCWMenubar.entitlements`.

## Prerequisites

- **Xcode 16+** - For building the menubar app
- **Go 1.21+** - For building the CLI (`ccw`)
- **iTerm2** (optional) - For terminal integration; falls back to Terminal.app if not installed

## How It Works

The CCW system has two components:
1. **CLI (`ccw`)** - Go binary that manages workspaces, git worktrees, and tmux sessions
2. **Menubar App** - SwiftUI app that provides a GUI and calls the CLI

The menubar app communicates with the CLI via subprocess execution. During development, use `CCW_BIN_PATH` to point to your local CLI build.

## Development Notes

### Code Organization
Most development happens in `CCWMenubarPackage/Sources/CCWMenubarFeature/` - organize your code as you prefer.

### Local Testing
Use the helper script to build and relaunch the menubar app locally:
```bash
scripts/dev-menubar.sh --release
```

Useful options and environment variables:
```bash
# Build debug configuration
scripts/dev-menubar.sh

# Skip Xcode build (relaunch only)
scripts/dev-menubar.sh --no-build

# Stream app logs after launch
scripts/dev-menubar.sh --release --logs

# Override build output location
DERIVED_DATA=/tmp/ccw-menubar-derived scripts/dev-menubar.sh --release

# Skip rebuilding the CLI binary
scripts/dev-menubar.sh --release --no-cli
```

### Logging

The app uses structured log prefixes for easy filtering:

| Prefix | Component |
|--------|-----------|
| `CCWMenubar[ui]` | UI lifecycle events |
| `CCWMenubar[app-state]` | AppState changes, workspace operations |
| `CCWMenubar[cli]` | CLI command execution |
| `CCWMenubar[delegate]` | App delegate lifecycle |
| `CCWMenubar[exit]` | Termination and keep-alive events |

Stream all logs:
```bash
/usr/bin/log stream --style compact \
  --predicate 'process == "CCWMenubar" || subsystem == "com.justanotheratom.ccw-menubar"' \
  --info --level info
```

Filter by component (grep stderr since NSLog goes there):
```bash
scripts/dev-menubar.sh 2>&1 | grep "CCWMenubar\[cli\]"
scripts/dev-menubar.sh 2>&1 | grep "CCWMenubar\[app-state\]"
```

### Public API Requirements
Types exposed to the app target need `public` access:
```swift
public struct SettingsView: View {
    public init() {}
    
    public var body: some View {
        // Your view code
    }
}
```

### Adding Dependencies
Edit `CCWMenubarPackage/Package.swift` to add SPM dependencies:
```swift
dependencies: [
    .package(url: "https://github.com/example/SomePackage", from: "1.0.0")
],
targets: [
    .target(
        name: "CCWMenubarFeature",
        dependencies: ["SomePackage"]
    ),
]
```

### Test Structure
- **Unit Tests**: `CCWMenubarPackage/Tests/CCWMenubarFeatureTests/` (Swift Testing framework)
- **UI Tests**: `CCWMenubarUITests/` (XCUITest framework)
- **Test Plan**: `CCWMenubar.xctestplan` coordinates all tests

## Configuration

### XCConfig Build Settings
Build settings are managed through **XCConfig files** in `Config/`:
- `Config/Shared.xcconfig` - Common settings (bundle ID, versions, deployment target)
- `Config/Debug.xcconfig` - Debug-specific settings  
- `Config/Release.xcconfig` - Release-specific settings
- `Config/Tests.xcconfig` - Test-specific settings

### App Sandbox & Entitlements
The app uses Apple Events automation to focus iTerm windows. Update `Config/CCWMenubar.entitlements` if additional capabilities are needed.

## Installation

### Homebrew Cask
```bash
brew install --cask justanotheratom/ccw/ccw-menubar
```

The Homebrew tap is maintained at: https://github.com/justanotheratom/homebrew-ccw

### Direct Download
Download the latest DMG from GitHub Releases and drag `CCWMenubar.app` to `/Applications`.

## macOS-Specific Features

### Window Management
Add multiple windows and settings panels:
```swift
@main
struct CCWMenubarApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
        
        Settings {
            SettingsView()
        }
    }
}
```

### Asset Management
- **App-Level Assets**: `CCWMenubar/Assets.xcassets/` (app icon with multiple sizes, accent color)
- **Feature Assets**: Add `Resources/` folder to SPM package if needed

### SPM Package Resources
To include assets in your feature package:
```swift
.target(
    name: "CCWMenubarFeature",
    dependencies: [],
    resources: [.process("Resources")]
)
```

## Releasing a New Version

### Step 1: Bump Version Numbers
Update version in two places:

```bash
# CLI version (cmd/root.go)
version = "X.Y.Z"

# Menubar app version (Config/Shared.xcconfig)
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
Pushing the tag triggers GitHub Actions:
- **release.yml** - Builds CLI binaries (darwin/linux, amd64/arm64), creates GitHub Release
- **release-menubar.yml** - Builds macOS app, signs/notarizes, creates DMG, uploads to release, updates Homebrew tap

### Step 4: Verify
```bash
gh release view vX.Y.Z
```

The Homebrew tap at https://github.com/justanotheratom/homebrew-ccw is automatically updated.

## Troubleshooting

### Menubar icon not appearing
The app may have crashed or failed to initialize. Check for zombie processes:
```bash
pgrep -fl CCWMenubar
```

Kill all instances and restart:
```bash
pkill -9 -f CCWMenubar
scripts/dev-menubar.sh
```

### CLI commands failing
Verify the CLI is accessible and working:
```bash
# Check if ccw is in PATH
which ccw

# Or use the local build
./bin/ccw version
./bin/ccw check --json
```

### Clean build
If you encounter strange build issues:
```bash
# Clean Xcode derived data
rm -rf .build/DerivedData

# Clean SPM build artifacts
rm -rf CCWMenubarPackage/.build

# Full rebuild
scripts/dev-menubar.sh
```

### Multiple instances running
The dev script should kill existing instances, but if you have duplicates:
```bash
pkill -9 -f CCWMenubar
```

### Workspace operations not working
Check the CLI logs for errors:
```bash
scripts/dev-menubar.sh 2>&1 | grep -E "CCWMenubar\[(cli|app-state)\]"
```

Common issues:
- **"workspace not found"** - Workspace ID format is `repo/branch` (e.g., `myapp/feature-x`)
- **iTerm not opening** - Ensure iTerm2 is installed, or the app will fall back to Terminal.app

## Notes

### Generated with XcodeBuildMCP
This project was scaffolded using [XcodeBuildMCP](https://github.com/cameroncooke/XcodeBuildMCP), which provides tools for AI-assisted macOS development workflows.
