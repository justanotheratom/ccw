import SwiftUI
import CCWMenubarFeature

@main
struct CCWMenubarApp: App {
    @StateObject private var appState = AppState()
    @State private var isMenuPresented = false

    var body: some Scene {
        MenuBarExtra("CCW", systemImage: statusImageName, isPresented: $isMenuPresented) {
            MenuBarView()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.window)
        .onChange(of: isMenuPresented) { newValue in
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
