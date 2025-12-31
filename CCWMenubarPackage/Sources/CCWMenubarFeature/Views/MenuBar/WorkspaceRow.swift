import SwiftUI

struct WorkspaceRow: View {
    let workspace: WorkspaceStatus
    @EnvironmentObject private var appState: AppState
    @State private var isHovered = false
    @State private var isConfirmingDelete = false

    private var isConnected: Bool {
        workspace.state == .connected
    }

    var body: some View {
        ZStack {
            // Normal state
            if !isConfirmingDelete {
                normalContent
                    .transition(.asymmetric(
                        insertion: .opacity.combined(with: .move(edge: .leading)),
                        removal: .opacity.combined(with: .move(edge: .leading))
                    ))
            }

            // Delete confirmation state
            if isConfirmingDelete {
                deleteConfirmContent
                    .transition(.asymmetric(
                        insertion: .opacity.combined(with: .move(edge: .trailing)),
                        removal: .opacity.combined(with: .move(edge: .trailing))
                    ))
            }
        }
        .animation(.easeInOut(duration: 0.2), value: isConfirmingDelete)
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .contentShape(Rectangle())
        .background(backgroundFill)
        .onHover { hovering in
            withAnimation(.easeInOut(duration: 0.15)) {
                isHovered = hovering
                if !hovering {
                    isConfirmingDelete = false
                }
            }
        }
    }

    // MARK: - Normal Content

    private var normalContent: some View {
        HStack(spacing: 8) {
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

            // Action buttons (shown on hover)
            if isHovered {
                HStack(spacing: 4) {
                    // Connect or Reveal button
                    Button(action: primaryAction) {
                        Image(systemName: isConnected ? "macwindow" : "play.fill")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundStyle(isConnected ? .blue : .green)
                            .frame(width: 24, height: 24)
                            .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .help(isConnected ? "Reveal window" : "Connect")

                    // Delete button
                    Button(action: { isConfirmingDelete = true }) {
                        Image(systemName: "trash")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundStyle(.secondary)
                            .frame(width: 24, height: 24)
                            .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .help("Delete workspace")
                }
            } else {
                // Time ago (shown when not hovered)
                Text(timeAgo)
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }
        }
        .onTapGesture {
            primaryAction()
        }
        .contextMenu {
            if isConnected {
                Button {
                    Task { await appState.openWorkspace(workspace.id) }
                } label: {
                    Label("Reveal", systemImage: "macwindow")
                }
            } else {
                Button {
                    Task { await appState.openWorkspace(workspace.id) }
                } label: {
                    Label("Connect", systemImage: "play.fill")
                }
            }

            Divider()

            Button(role: .destructive) {
                isConfirmingDelete = true
            } label: {
                Label("Delete", systemImage: "trash")
            }
        }
    }

    // MARK: - Delete Confirmation Content

    private var deleteConfirmContent: some View {
        HStack(spacing: 12) {
            Image(systemName: "trash.fill")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(.red)

            Text("Delete?")
                .font(.system(size: 13, weight: .medium))
                .foregroundStyle(.primary)

            Spacer()

            HStack(spacing: 8) {
                Button("Cancel") {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        isConfirmingDelete = false
                    }
                }
                .buttonStyle(.plain)
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(.secondary)

                Button("Delete") {
                    deleteWorkspace()
                }
                .buttonStyle(.plain)
                .font(.system(size: 12, weight: .semibold))
                .foregroundStyle(.white)
                .padding(.horizontal, 12)
                .padding(.vertical, 4)
                .background(
                    RoundedRectangle(cornerRadius: 4)
                        .fill(.red)
                )
            }
        }
    }

    // MARK: - Background

    private var backgroundFill: some View {
        Group {
            if isConfirmingDelete {
                Color.red.opacity(0.08)
            } else if isHovered {
                Color.primary.opacity(0.05)
            } else {
                Color.clear
            }
        }
    }

    // MARK: - Actions

    private func primaryAction() {
        Task { await appState.openWorkspace(workspace.id) }
    }

    private func deleteWorkspace() {
        let id = workspace.id
        Task {
            await appState.removeWorkspace(id, force: false)
        }
    }

    // MARK: - Computed Properties

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
