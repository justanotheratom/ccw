import SwiftUI
import CCWMenubarFeature
import KeyboardShortcuts

@main
struct CCWMenubarApp: App {
    @StateObject private var appState = AppState()
    @StateObject private var menuState: MenuState
    private let logger = CCWLog.ui

    init() {
        let state = MenuState()
        _menuState = StateObject(wrappedValue: state)
        KeyboardShortcuts.onKeyUp(for: .toggleMenu) {
            DispatchQueue.main.async {
                state.isInserted.toggle()
            }
        }
    }

    var body: some Scene {
        MenuBarExtra("CCW", systemImage: statusImageName, isInserted: $menuState.isInserted) {
            MenuBarView()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.window)
        .onChange(of: menuState.isInserted) { newValue in
            logger.info("menu bar extra isInserted=\(newValue, privacy: .public)")
            if newValue {
                Task { await appState.refreshWorkspaces() }
            }
        }

        Settings {
            SettingsView()
                .environmentObject(appState)
        }
    }

    private var statusImageName: String {
        if appState.workspaces.contains(where: { $0.state == .connected }) {
            return "circle.fill"
        }
        if appState.workspaces.contains(where: { $0.state == .alive }) {
            return "circle"
        }
        if appState.workspaces.contains(where: { $0.state == .dead }) {
            return "xmark.circle"
        }
        return "terminal.fill"
    }
}

final class MenuState: ObservableObject {
    @Published var isInserted = true
}
