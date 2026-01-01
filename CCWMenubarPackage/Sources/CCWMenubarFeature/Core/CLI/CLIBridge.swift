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
        let env = ProcessInfo.processInfo.environment
        if let override = env["CCW_BIN_PATH"], FileManager.default.isExecutableFile(atPath: override) {
            self.ccwURL = URL(fileURLWithPath: override)
            logger.notice("using CCW_BIN_PATH override at \(override, privacy: .public)")
            NSLog("CCWMenubar[cli] using CCW_BIN_PATH override at \(override)")
        } else if let url = Bundle.main.url(forResource: "ccw", withExtension: nil) {
            self.ccwURL = url
        } else if let url = Self.findCCWInPath() {
            self.ccwURL = url
            logger.notice("ccw not in bundle, falling back to PATH at \(url.path, privacy: .public)")
            NSLog("CCWMenubar[cli] ccw not in bundle, falling back to PATH at \(url.path)")
        } else {
            throw CLIError.ccwNotFound
        }
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let value = try container.decode(String.self)
            if let date = Self.parseISO8601Date(value) {
                return date
            }
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Invalid ISO8601 date: \(value)")
        }
        self.decoder = decoder
    }

    public func listWorkspaces() async throws -> [WorkspaceStatus] {
        let output = try await execute(["ls", "--json"])
        if output.isEmpty {
            return []
        }
        return try decoder.decode([WorkspaceStatus]?.self, from: output) ?? []
    }

    public func openWorkspace(_ id: String, resume: Bool = true, focusExisting: Bool = true, forceAttach: Bool = true) async throws {
        var args = ["open", id]
        if !resume {
            args.append("--no-resume")
        }
        if focusExisting {
            args.append("--focus")
        }
        if forceAttach {
            args.append("--attach")
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

        let stdoutHandle = stdout.fileHandleForReading
        let stderrHandle = stderr.fileHandleForReading
        let stdoutTask = Task { try await stdoutHandle.readToEnd() ?? Data() }
        let stderrTask = Task { try await stderrHandle.readToEnd() ?? Data() }

        let status: Int32
        do {
            status = try await withCheckedThrowingContinuation { continuation in
                process.terminationHandler = { process in
                    continuation.resume(returning: process.terminationStatus)
                }
                do {
                    try process.run()
                } catch {
                    process.terminationHandler = nil
                    continuation.resume(throwing: error)
                }
            }
        } catch {
            stdoutTask.cancel()
            stderrTask.cancel()
            throw error
        }
        process.terminationHandler = nil

        let output = try await stdoutTask.value
        let errorData = try await stderrTask.value
        if status != 0 {
            let errorMsg = String(data: errorData, encoding: .utf8) ?? "Unknown error"
            let elapsed = Date().timeIntervalSince(start)
            logger.error("execute failed status=\(status, privacy: .public) elapsed=\(elapsed, privacy: .public)s error=\(errorMsg, privacy: .public)")
            NSLog("CCWMenubar[cli] execute failed status=\(status) elapsed=\(elapsed)s error=\(errorMsg)")
            throw CLIError.commandFailed(errorMsg)
        }

        let elapsed = Date().timeIntervalSince(start)
        logger.info("execute success elapsed=\(elapsed, privacy: .public)s bytes=\(output.count, privacy: .public)")
        NSLog("CCWMenubar[cli] execute success elapsed=\(elapsed)s bytes=\(output.count)")
        return output
    }

    private static func findCCWInPath() -> URL? {
#if DEBUG
        let env = ProcessInfo.processInfo.environment
        let defaultPath = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
        let envPath = env["PATH"] ?? ""
        let combined = [envPath, defaultPath].joined(separator: ":")
        for entry in combined.split(separator: ":") {
            let candidate = URL(fileURLWithPath: String(entry)).appendingPathComponent("ccw")
            if FileManager.default.isExecutableFile(atPath: candidate.path) {
                return candidate
            }
        }
#endif
        return nil
    }

    private static func parseISO8601Date(_ value: String) -> Date? {
        let fractionalFormatter = ISO8601DateFormatter()
        fractionalFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = fractionalFormatter.date(from: value) {
            return date
        }
        let fallbackFormatter = ISO8601DateFormatter()
        return fallbackFormatter.date(from: value)
    }
}
