import SwiftUI

public struct SettingsView: View {
    @EnvironmentObject private var appState: AppState

    @State private var reposDir = ""
    @State private var layoutLeft = "claude"
    @State private var layoutRight = "lazygit"
    @State private var itermCCMode = false
    @State private var skipPerms = false
    @State private var showingOnboarding = false

    public init() {}

    public var body: some View {
        Form {
            HStack {
                Text("Repos Directory")
                TextField("", text: $reposDir)
            }

            VStack(alignment: .leading) {
                Text("Layout")
                HStack {
                    TextField("Left", text: $layoutLeft)
                    TextField("Right", text: $layoutRight)
                }
            }

            Toggle("iTerm CC Mode", isOn: $itermCCMode)
            Toggle("Skip permission prompts", isOn: $skipPerms)

            HStack {
                Button("Re-run Setup") {
                    Task {
                        await appState.setConfig(key: "onboarded", value: "false")
                        appState.setupState = .needsOnboarding
                        showingOnboarding = true
                    }
                }
                Spacer()
                Button("Save") {
                    Task { await saveConfig() }
                }
            }
        }
        .padding()
        .frame(width: 520)
        .task { await loadConfig() }
        .sheet(isPresented: $showingOnboarding) {
            OnboardingView()
                .environmentObject(appState)
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
