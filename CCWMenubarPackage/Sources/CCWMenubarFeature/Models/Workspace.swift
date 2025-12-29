import Foundation

public struct Workspace: Codable, Sendable {
    public let repo: String
    public let repoPath: String
    public let branch: String
    public let baseBranch: String
    public let worktreePath: String
    public let claudeSession: String
    public let tmuxSession: String
    public let createdAt: Date
    public let lastAccessedAt: Date

    public enum CodingKeys: String, CodingKey {
        case repo
        case repoPath = "repo_path"
        case branch
        case baseBranch = "base_branch"
        case worktreePath = "worktree_path"
        case claudeSession = "claude_session"
        case tmuxSession = "tmux_session"
        case createdAt = "created_at"
        case lastAccessedAt = "last_accessed_at"
    }
}

public struct WorkspaceStatus: Codable, Identifiable, Sendable {
    public let id: String
    public let workspace: Workspace
    public let sessionAlive: Bool
    public let hasClients: Bool

    public enum CodingKeys: String, CodingKey {
        case id = "ID"
        case workspace = "Workspace"
        case sessionAlive = "SessionAlive"
        case hasClients = "HasClients"
    }

    public var state: WorkspaceState {
        if !sessionAlive { return .dead }
        if hasClients { return .connected }
        return .alive
    }
}

public enum WorkspaceState {
    case connected
    case alive
    case dead
}

public typealias WorkspaceInfo = WorkspaceStatus
