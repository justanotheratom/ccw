import Testing
@testable import CCWMenubarFeature

private func makeDecoder() -> JSONDecoder {
    let decoder = JSONDecoder()
    decoder.dateDecodingStrategy = .iso8601
    return decoder
}

@Test func decodeWorkspaceStatus() throws {
    let json = """
    {
      "ID": "repo/branch",
      "Workspace": {
        "repo": "repo",
        "repo_path": "/Users/me/dev/repo",
        "branch": "branch",
        "base_branch": "main",
        "worktree_path": "/Users/me/dev/repo-branch",
        "claude_session": "ccw-repo-branch",
        "tmux_session": "ccw-repo-branch",
        "created_at": "2024-01-01T12:00:00Z",
        "last_accessed_at": "2024-01-10T09:30:00Z"
      },
      "SessionAlive": true,
      "HasClients": false
    }
    """
    let data = Data(json.utf8)
    let status = try makeDecoder().decode(WorkspaceStatus.self, from: data)
    #expect(status.id == "repo/branch")
    #expect(status.workspace.repoPath == "/Users/me/dev/repo")
    #expect(status.sessionAlive == true)
    #expect(status.hasClients == false)
}

@Test func decodeStaleList() throws {
    let json = """
    [
      {
        "ID": "repo/branch",
        "Workspace": {
          "repo": "repo",
          "repo_path": "/Users/me/dev/repo",
          "branch": "branch",
          "base_branch": "main",
          "worktree_path": "/Users/me/dev/repo-branch",
          "claude_session": "ccw-repo-branch",
          "tmux_session": "ccw-repo-branch",
          "created_at": "2024-01-01T12:00:00Z",
          "last_accessed_at": "2024-01-10T09:30:00Z"
        },
        "SessionAlive": false,
        "HasClients": false
      }
    ]
    """
    let data = Data(json.utf8)
    let statuses = try makeDecoder().decode([WorkspaceStatus].self, from: data)
    #expect(statuses.count == 1)
    #expect(statuses[0].state == .dead)
}

@Test func decodeConfig() throws {
    let json = """
    {
      "version": 1,
      "repos_dir": "~/dev",
      "iterm_cc_mode": true,
      "claude_rename_delay": 0,
      "layout": { "left": "claude", "right": "lazygit" },
      "onboarded": true,
      "claude_dangerously_skip_permissions": false
    }
    """
    let data = Data(json.utf8)
    let config = try makeDecoder().decode(CCWConfig.self, from: data)
    #expect(config.reposDir == "~/dev")
    #expect(config.layout.left == "claude")
}

@Test func decodeDependencyStatus() throws {
    let json = """
    {
      "git": { "installed": true, "path": "/usr/bin/git" },
      "iterm": { "installed": true, "path": "/Applications/iTerm.app" },
      "lazygit": { "installed": false, "path": "", "optional": true }
    }
    """
    let data = Data(json.utf8)
    let deps = try makeDecoder().decode([String: DepStatus].self, from: data)
    #expect(deps["git"]?.installed == true)
    #expect(deps["lazygit"]?.optional == true)
}
