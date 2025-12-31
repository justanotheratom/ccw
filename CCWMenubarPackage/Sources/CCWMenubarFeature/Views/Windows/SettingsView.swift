import SwiftUI
import KeyboardShortcuts
import ServiceManagement

public struct SettingsView: View {
    @EnvironmentObject private var appState: AppState
    @Binding var navigationPath: NavigationPath

    @StateObject private var launchAtLogin = LaunchAtLoginModel()
    @State private var reposDir = ""
    @State private var layoutLeft = "claude"
    @State private var layoutRight = "lazygit"
    @State private var itermCCMode = false
    @State private var skipPerms = false

    public init(navigationPath: Binding<NavigationPath>) {
        self._navigationPath = navigationPath
    }

    private var versionString: String {
        let version = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "unknown"
        let build = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "0"
        return "Version \(version) (\(build))"
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                // Repositories Section
                SettingsSection(title: "REPOSITORIES") {
                    VStack(alignment: .leading, spacing: 6) {
                        Text("Directory")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundStyle(.secondary)
                        TextField("~/github", text: $reposDir)
                            .textFieldStyle(.roundedBorder)
                            .font(.system(size: 13))
                    }
                }

                // Layout Section
                SettingsSection(title: "LAYOUT") {
                    HStack(spacing: 12) {
                        VStack(alignment: .leading, spacing: 6) {
                            Text("Left Pane")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundStyle(.secondary)
                            TextField("claude", text: $layoutLeft)
                                .textFieldStyle(.roundedBorder)
                                .font(.system(size: 13, design: .monospaced))
                        }

                        VStack(alignment: .leading, spacing: 6) {
                            Text("Right Pane")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundStyle(.secondary)
                            TextField("lazygit", text: $layoutRight)
                                .textFieldStyle(.roundedBorder)
                                .font(.system(size: 13, design: .monospaced))
                        }
                    }
                }

                // Behavior Section
                SettingsSection(title: "BEHAVIOR") {
                    VStack(alignment: .leading, spacing: 10) {
                        Toggle("iTerm CC Mode", isOn: $itermCCMode)
                            .font(.system(size: 13))

                        Toggle("Skip Permission Prompts", isOn: $skipPerms)
                            .font(.system(size: 13))

                        Toggle("Launch at Login", isOn: Binding(
                            get: { launchAtLogin.isEnabled },
                            set: { launchAtLogin.setEnabled($0) }
                        ))
                        .font(.system(size: 13))
                    }
                }

                // Keyboard Section
                SettingsSection(title: "KEYBOARD") {
                    HStack {
                        Text("Toggle Menu")
                            .font(.system(size: 13))
                        Spacer()
                        KeyboardShortcuts.Recorder("", name: .toggleMenu)
                    }
                }

                // Actions Section
                SettingsSection(title: "ACTIONS") {
                    Button("Re-run Setup") {
                        Task {
                            await appState.setConfig(key: "onboarded", value: "false")
                            appState.setupState = .needsOnboarding
                            navigationPath.removeLast()
                        }
                    }
                    .font(.system(size: 13))
                }
            }
            .padding(16)
        }
        .safeAreaInset(edge: .bottom) {
            VStack(spacing: 0) {
                Divider()
                    .opacity(0.3)

                HStack {
                    Text(versionString)
                        .font(.system(size: 11))
                        .foregroundStyle(.tertiary)

                    Spacer()

                    Button("Cancel") {
                        navigationPath.removeLast()
                    }
                    .keyboardShortcut(.cancelAction)

                    Button("Save") {
                        Task {
                            await saveConfig()
                            navigationPath.removeLast()
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut(.defaultAction)
                }
                .padding(16)
            }
            .background(.ultraThinMaterial)
        }
        .frame(width: 320)
        .background(.ultraThinMaterial)
        .navigationTitle("Settings")
        .task {
            await loadConfig()
            await launchAtLogin.refresh()
        }
        .onAppear {
            NSLog("CCWMenubar[ui] settings version=\(versionString)")
        }
    }

    private func loadConfig() async {
        await appState.loadConfig()
        guard let config = appState.config else { return }
        reposDir = config.reposDir
        layoutLeft = config.layout.left
        layoutRight = config.layout.right
        itermCCMode = config.itermCCMode
        skipPerms = config.claudeDangerouslySkipPermissions
    }

    private func saveConfig() async {
        await appState.setConfig(key: "repos_dir", value: reposDir)
        await appState.setConfig(key: "layout.left", value: layoutLeft)
        await appState.setConfig(key: "layout.right", value: layoutRight)
        await appState.setConfig(key: "iterm_cc_mode", value: itermCCMode ? "true" : "false")
        await appState.setConfig(key: "claude_dangerously_skip_permissions", value: skipPerms ? "true" : "false")
        await appState.loadConfig()
    }
}

// MARK: - Settings Section Component

private struct SettingsSection<Content: View>: View {
    let title: String
    @ViewBuilder let content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title)
                .font(.system(size: 10, weight: .semibold))
                .foregroundStyle(.secondary)
                .tracking(0.5)

            content
        }
    }
}

// MARK: - Keyboard Shortcuts Extension

extension KeyboardShortcuts.Name {
    public static let toggleMenu = Self("toggleMenu")
}

// MARK: - Launch at Login Model

@MainActor
final class LaunchAtLoginModel: ObservableObject {
    @Published var isEnabled = false
    @Published var lastError: String?

    private let logger = CCWLog.ui

    func refresh() async {
        await Task.yield()
        let status = SMAppService.mainApp.status
        isEnabled = (status == .enabled)
        logger.notice("launch-at-login refresh status=\(String(describing: status), privacy: .public)")
        NSLog("CCWMenubar[ui] launch-at-login refresh status=\(status)")
    }

    func setEnabled(_ newValue: Bool) {
        Task { await updateEnabled(newValue) }
    }

    private func updateEnabled(_ newValue: Bool) async {
        await Task.yield()
        logger.notice("launch-at-login toggle requested=\(newValue, privacy: .public)")
        NSLog("CCWMenubar[ui] launch-at-login toggle requested=\(newValue)")
        do {
            if newValue {
                try await SMAppService.mainApp.register()
            } else {
                try await SMAppService.mainApp.unregister()
            }
            lastError = nil
        } catch {
            lastError = error.localizedDescription
            logger.error("launch-at-login toggle failed error=\(error.localizedDescription, privacy: .public)")
            NSLog("CCWMenubar[ui] launch-at-login toggle failed error=\(error.localizedDescription)")
        }
        await refresh()
    }
}
