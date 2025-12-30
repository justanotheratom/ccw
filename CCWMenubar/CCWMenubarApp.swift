import SwiftUI
import CCWMenubarFeature
import KeyboardShortcuts

@main
struct CCWMenubarApp: App {
    @StateObject private var appState: AppState
    @StateObject private var menuState: MenuState
    private let logger = CCWLog.ui

    init() {
        let appState = AppState()
        _appState = StateObject(wrappedValue: appState)
        let state = MenuState()
        _menuState = StateObject(wrappedValue: state)
        KeyboardShortcuts.onKeyUp(for: .toggleMenu) {
            DispatchQueue.main.async {
                state.isInserted.toggle()
            }
        }
        DispatchQueue.main.async {
            appState.start()
        }
    }

    var body: some Scene {
        MenuBarExtra("CCW", systemImage: statusImageName, isInserted: isInsertedBinding) {
            MenuBarView()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.window)
        .onChange(of: menuState.isInserted) { newValue in
            logger.info("menu bar extra isInserted=\(newValue, privacy: .public)")
            NSLog("CCWMenubar[ui] menu bar extra isInserted=\(newValue)")
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

    private var isInsertedBinding: Binding<Bool> {
        Binding(
            get: { menuState.isInserted },
            set: { newValue in
                if menuState.isInserted != newValue {
                    menuState.isInserted = newValue
                }
            }
        )
    }
}

final class MenuState: ObservableObject {
    @Published var isInserted = true {
        didSet {
            NSLog("CCWMenubar[ui] menuState isInserted didSet value=\(isInserted)")
        }
    }
}
