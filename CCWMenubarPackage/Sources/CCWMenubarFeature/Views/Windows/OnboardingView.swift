import SwiftUI

struct OnboardingView: View {
    @EnvironmentObject private var appState: AppState
    @Environment(\.dismiss) private var dismiss

    @State private var reposDir = ""
    @State private var layoutLeft = "claude"
    @State private var layoutRight = "lazygit"
    @State private var skipPerms = false
    @State private var deps: [String: DepStatus] = [:]

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Welcome to CCW Menubar")
                .font(.headline)

            VStack(alignment: .leading, spacing: 4) {
                Text("Dependencies")
                ForEach(deps.keys.sorted(), id: \.self) { key in
                    let dep = deps[key]
                    HStack {
                        Text(key)
                        Spacer()
                        Text(dep?.installed == true ? "installed" : "missing")
                            .foregroundColor(dep?.installed == true ? .green : .red)
                    }
                }
                Text("If iTerm focus fails, allow Automation for CCW Menubar in System Settings.")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }

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

            Toggle("Skip Claude permission prompts", isOn: $skipPerms)

            HStack {
                Button("Cancel") { dismiss() }
                Spacer()
                Button("Complete Setup") {
                    Task { await completeSetup() }
                }
                .disabled(reposDir.isEmpty || missingRequiredDeps)
            }
        }
        .padding()
        .frame(width: 520)
        .task { await load() }
    }

    private var missingRequiredDeps: Bool {
        deps.contains { !$0.value.installed && $0.value.optional != true }
    }

    private func load() async {
        deps = await appState.checkDependencies()
        await appState.loadConfig()
        if let config = appState.config {
            reposDir = config.reposDir
            layoutLeft = config.layout.left
            layoutRight = config.layout.right
            skipPerms = config.claudeDangerouslySkipPermissions
        }
    }

    private func completeSetup() async {
        await appState.setConfig(key: "repos_dir", value: reposDir)
        await appState.setConfig(key: "layout.left", value: layoutLeft)
        await appState.setConfig(key: "layout.right", value: layoutRight)
        await appState.setConfig(key: "claude_dangerously_skip_permissions", value: skipPerms ? "true" : "false")
        await appState.setConfig(key: "onboarded", value: "true")
        await appState.loadConfig()
        appState.setupState = .ready
        await appState.refreshWorkspaces()
        dismiss()
    }
}
