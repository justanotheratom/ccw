import SwiftUI

struct NewWorkspaceView: View {
    @EnvironmentObject private var appState: AppState
    @Binding var navigationPath: NavigationPath

    @State private var repos: [String] = []
    @State private var selectedRepo = ""
    @State private var branchName = ""
    @State private var baseBranch = "main"
    @State private var isCreating = false

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Form content
            VStack(alignment: .leading, spacing: 16) {
                // Repository picker
                VStack(alignment: .leading, spacing: 6) {
                    Text("Repository")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundStyle(.secondary)

                    Picker("", selection: $selectedRepo) {
                        if repos.isEmpty {
                            Text("Loading...").tag("")
                        } else {
                            ForEach(repos, id: \.self) { repo in
                                Text(repo).tag(repo)
                            }
                        }
                    }
                    .pickerStyle(.menu)
                    .labelsHidden()
                }

                // Branch name
                VStack(alignment: .leading, spacing: 6) {
                    Text("Branch Name")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundStyle(.secondary)

                    TextField("feature/my-feature", text: $branchName)
                        .textFieldStyle(.roundedBorder)
                        .font(.system(size: 13, design: .monospaced))
                }

                // Base branch
                VStack(alignment: .leading, spacing: 6) {
                    Text("Base Branch")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundStyle(.secondary)

                    TextField("main", text: $baseBranch)
                        .textFieldStyle(.roundedBorder)
                        .font(.system(size: 13, design: .monospaced))
                }
            }
            .padding(16)

            Divider()
                .opacity(0.3)

            // Action buttons
            HStack {
                Button("Cancel") {
                    navigationPath.removeLast()
                }
                .keyboardShortcut(.cancelAction)

                Spacer()

                Button(action: create) {
                    if isCreating {
                        ProgressView()
                            .controlSize(.small)
                            .frame(width: 60)
                    } else {
                        Text("Create")
                            .frame(width: 60)
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(selectedRepo.isEmpty || branchName.isEmpty || isCreating)
                .keyboardShortcut(.defaultAction)
            }
            .padding(16)
        }
        .frame(width: 320)
        .background(.ultraThinMaterial)
        .navigationTitle("New Workspace")
        .task { await loadRepos() }
    }

    private func loadRepos() async {
        repos = await appState.listRepos()
        if selectedRepo.isEmpty {
            selectedRepo = repos.first ?? ""
        }
    }

    private func create() {
        isCreating = true
        Task {
            await appState.createWorkspace(repo: selectedRepo, branch: branchName, base: baseBranch.isEmpty ? nil : baseBranch)
            // Open the newly created workspace
            let workspaceId = "\(selectedRepo)/\(branchName)"
            await appState.openWorkspace(workspaceId)
            isCreating = false
            navigationPath.removeLast()
        }
    }
}
