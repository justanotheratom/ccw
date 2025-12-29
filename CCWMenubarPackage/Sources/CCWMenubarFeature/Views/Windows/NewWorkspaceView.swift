import SwiftUI

struct NewWorkspaceView: View {
    @EnvironmentObject private var appState: AppState
    @Environment(\.dismiss) private var dismiss

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
                    .disabled(selectedRepo.isEmpty || branchName.isEmpty || isCreating)
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding()
        .frame(width: 400)
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
            isCreating = false
            dismiss()
        }
    }
}
