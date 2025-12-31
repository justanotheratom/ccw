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

Log inspection without the script:
```bash
/usr/bin/log stream --style compact \
  --predicate 'process == "CCWMenubar" || subsystem == "com.justanotheratom.ccw-menubar"' \
  --info --level info
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
brew install --cask justanotheratom/tap/ccw-menubar
```

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

## Notes

### Generated with XcodeBuildMCP
This project was scaffolded using [XcodeBuildMCP](https://github.com/cameroncooke/XcodeBuildMCP), which provides tools for AI-assisted macOS development workflows.
