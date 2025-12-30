import Foundation

public actor CLIBridge {
    private let logger = CCWLog.cli
    public enum CLIError: Error, LocalizedError {
        case commandFailed(String)
        case ccwNotFound

        public var errorDescription: String? {
            switch self {
            case .commandFailed(let msg):
                return "Command failed: \(msg)"
            case .ccwNotFound:
                return "ccw binary not found in app bundle"
            }
        }
    }

    private let ccwURL: URL
    private let decoder: JSONDecoder

    public init() throws {
        guard let url = Bundle.main.url(forAuxiliaryExecutable: "ccw") else {
            throw CLIError.ccwNotFound
        }
        self.ccwURL = url
        let decoder = JSONDecoder()
        let fractionalFormatter = ISO8601DateFormatter()
        fractionalFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let fallbackFormatter = ISO8601DateFormatter()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let value = try container.decode(String.self)
            if let date = fractionalFormatter.date(from: value) {
                return date
            }
            if let date = fallbackFormatter.date(from: value) {
                return date
            }
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Invalid ISO8601 date: \(value)")
        }
        self.decoder = decoder
    }

    public func listWorkspaces() async throws -> [WorkspaceStatus] {
        let output = try await execute(["ls", "--json"])
        return try decoder.decode([WorkspaceStatus].self, from: output)
    }

    public func openWorkspace(_ id: String, resume: Bool = true) async throws {
        var args = ["open", id]
        if !resume {
            args.append("--no-resume")
        }
        _ = try await execute(args)
    }

    public func createWorkspace(repo: String, branch: String, base: String?) async throws {
        var args = ["new", repo, branch, "--no-attach"]
        if let base = base {
            args.append(contentsOf: ["--base", base])
        }
        _ = try await execute(args)
    }

    public func removeWorkspace(_ id: String, force: Bool) async throws {
        var args = ["rm", id, "--yes"]
        if force {
            args.append("--force")
        }
        _ = try await execute(args)
    }

    public func staleWorkspaces() async throws -> [WorkspaceStatus] {
        let output = try await execute(["stale", "--json"])
        return try decoder.decode([WorkspaceStatus]?.self, from: output) ?? []
    }

    public func workspaceInfo(_ id: String) async throws -> WorkspaceInfo {
        let output = try await execute(["info", id, "--json"])
        return try decoder.decode(WorkspaceInfo.self, from: output)
    }

    public func listRepos() async throws -> [String] {
        let output = try await execute(["repos", "--json"])
        return try decoder.decode([String].self, from: output)
    }

    public func checkDependencies() async throws -> [String: DepStatus] {
        let output = try await execute(["check", "--json"])
        return try decoder.decode([String: DepStatus].self, from: output)
    }

    public func getConfig() async throws -> CCWConfig {
        let configPath = FileManager.default.homeDirectoryForCurrentUser
            .appendingPathComponent(".ccw/config.json")
        let data = try Data(contentsOf: configPath)
        return try decoder.decode(CCWConfig.self, from: data)
    }

    public func setConfig(key: String, value: String) async throws {
        _ = try await execute(["config", key, value])
    }

    private func execute(_ arguments: [String]) async throws -> Data {
        let start = Date()
        let argString = arguments.joined(separator: " ")
        logger.info("execute start args=\(argString, privacy: .public)")
        NSLog("CCWMenubar[cli] execute start args=\(argString)")
        let process = Process()
        process.executableURL = ccwURL
        process.arguments = arguments

        var env = ProcessInfo.processInfo.environment
        env["PATH"] = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
        process.environment = env

        let stdout = Pipe()
        let stderr = Pipe()
        process.standardOutput = stdout
        process.standardError = stderr

        try process.run()
        process.waitUntilExit()

        if process.terminationStatus != 0 {
            let errorData = stderr.fileHandleForReading.readDataToEndOfFile()
            let errorMsg = String(data: errorData, encoding: .utf8) ?? "Unknown error"
            let elapsed = Date().timeIntervalSince(start)
            logger.error("execute failed status=\(process.terminationStatus, privacy: .public) elapsed=\(elapsed, privacy: .public)s error=\(errorMsg, privacy: .public)")
            NSLog("CCWMenubar[cli] execute failed status=\(process.terminationStatus) elapsed=\(elapsed)s error=\(errorMsg)")
            throw CLIError.commandFailed(errorMsg)
        }

        let output = stdout.fileHandleForReading.readDataToEndOfFile()
        let elapsed = Date().timeIntervalSince(start)
        logger.info("execute success elapsed=\(elapsed, privacy: .public)s bytes=\(output.count, privacy: .public)")
        NSLog("CCWMenubar[cli] execute success elapsed=\(elapsed)s bytes=\(output.count)")
        return output
    }
}
