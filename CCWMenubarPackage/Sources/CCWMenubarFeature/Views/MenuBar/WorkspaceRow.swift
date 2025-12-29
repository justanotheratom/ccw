import SwiftUI

struct WorkspaceRow: View {
    let workspace: WorkspaceStatus
    @EnvironmentObject private var appState: AppState
    @State private var confirmRemove = false

    var body: some View {
        Button(action: {
            Task { await appState.openWorkspace(workspace.id) }
        }) {
            HStack {
                StatusIndicator(state: workspace.state)
                Text(workspace.id)
                Spacer()
                Text(timeAgo)
                    .foregroundColor(.secondary)
                    .font(.caption)
            }
            .padding(.horizontal)
            .padding(.vertical, 6)
        }
        .buttonStyle(.plain)
        .contextMenu {
            Button("Open") {
                Task { await appState.openWorkspace(workspace.id) }
            }
            Button("Open (no resume)") {
                Task { await appState.openWorkspace(workspace.id, resume: false) }
            }
            Button("Info") {
                Task { await appState.loadWorkspaceInfo(workspace.id) }
            }
            Divider()
            Button("Remove") {
                confirmRemove = true
            }
        }
        .alert("Remove workspace?", isPresented: $confirmRemove) {
            Button("Remove", role: .destructive) {
                Task { await appState.removeWorkspace(workspace.id, force: false) }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("This will remove the worktree and session for \(workspace.id).")
        }
    }

    private var timeAgo: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: workspace.workspace.lastAccessedAt, relativeTo: Date())
    }
}
