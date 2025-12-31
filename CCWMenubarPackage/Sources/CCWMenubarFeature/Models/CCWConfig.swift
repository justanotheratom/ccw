import Foundation

public struct CCWConfig: Codable, Sendable {
    public var version: Int
    public var reposDir: String
    public var itermCCMode: Bool
    public var claudeRenameDelay: Int
    public var layout: Layout
    public var onboarded: Bool
    public var claudeDangerouslySkipPermissions: Bool

    public struct Layout: Codable, Sendable {
        public var left: String
        public var right: String
    }

    public enum CodingKeys: String, CodingKey {
        case version
        case reposDir = "repos_dir"
        case itermCCMode = "iterm_cc_mode"
        case claudeRenameDelay = "claude_rename_delay"
        case layout
        case onboarded
        case claudeDangerouslySkipPermissions = "claude_dangerously_skip_permissions"
    }
}
