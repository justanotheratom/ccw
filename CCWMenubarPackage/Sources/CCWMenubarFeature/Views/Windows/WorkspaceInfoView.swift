import SwiftUI

struct WorkspaceInfoView: View {
    let info: WorkspaceInfo?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            if let info {
                Text(info.id)
                    .font(.headline)
                Text("Repo: \(info.workspace.repoPath)")
                Text("Worktree: \(info.workspace.worktreePath)")
                Text("Branch: \(info.workspace.branch)")
                Text("Base: \(info.workspace.baseBranch)")
                Text("Claude Session: \(info.workspace.claudeSession)")
                Text("Tmux Session: \(info.workspace.tmuxSession)")
                Text("Session Alive: \(info.sessionAlive ? "true" : "false")")
            } else {
                Text("No workspace info available.")
            }
        }
        .padding()
        .frame(width: 420)
    }
}
