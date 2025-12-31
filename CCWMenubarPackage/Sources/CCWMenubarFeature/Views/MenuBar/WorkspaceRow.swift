import SwiftUI

struct WorkspaceRow: View {
    let workspace: WorkspaceStatus
    @EnvironmentObject private var appState: AppState
    @State private var confirmRemove = false
    @State private var isHovered = false

    var body: some View {
        Button(action: {
            Task { await appState.openWorkspace(workspace.id) }
        }) {
            HStack(spacing: 12) {
                // Status indicator with glow
                Circle()
                    .fill(statusColor)
                    .frame(width: 10, height: 10)
                    .shadow(color: statusColor.opacity(0.5), radius: 4)

                // Workspace ID in monospace
                Text(workspace.id)
                    .font(.system(size: 13, design: .monospaced))
                    .lineLimit(1)
                    .truncationMode(.middle)

                Spacer()

                // Time ago
                Text(timeAgo)
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 10)
            .contentShape(Rectangle())
            .background(isHovered ? Color.primary.opacity(0.05) : Color.clear)
        }
        .buttonStyle(.plain)
        .onHover { hovering in
            isHovered = hovering
        }
        .contextMenu {
            Button {
                Task { await appState.openWorkspace(workspace.id) }
            } label: {
                Label("Open", systemImage: "arrow.up.forward.app")
            }

            Button {
                Task { await appState.openWorkspace(workspace.id, resume: false) }
            } label: {
                Label("Open (no resume)", systemImage: "arrow.clockwise")
            }

            Button {
                Task { await appState.loadWorkspaceInfo(workspace.id) }
            } label: {
                Label("Info", systemImage: "info.circle")
            }

            Divider()

            Button(role: .destructive) {
                confirmRemove = true
            } label: {
                Label("Remove", systemImage: "trash")
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

    private var statusColor: Color {
        switch workspace.state {
        case .connected: return .green
        case .alive: return .yellow
        case .dead: return .gray
        }
    }

    private var timeAgo: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: workspace.workspace.lastAccessedAt, relativeTo: Date())
    }
}
