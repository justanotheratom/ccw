import Foundation
import SwiftUI

@MainActor
public final class AppState: ObservableObject {
    private let logger = CCWLog.appState
    @Published public var workspaces: [WorkspaceStatus] = []
    @Published public var staleWorkspaces: [WorkspaceStatus] = []
    @Published public var isLoading = false
    @Published public var error: Error?
    @Published public var workspaceInfo: WorkspaceInfo?
    @Published public var showingWorkspaceInfo = false
    @Published public var config: CCWConfig?
    @Published public var setupState: SetupState = .checking

    public enum SetupState {
        case checking
        case needsOnboarding
        case missingDependencies([String: DepStatus])
        case ready
        case error(String)
    }

    private var cli: CLIBridge?

    public init() {
        logger.info("init start (mainThread=\(Thread.isMainThread, privacy: .public))")
        Task { await initialize() }
    }

    private func initialize() async {
        do {
            logger.info("initialize: creating CLI bridge")
            cli = try CLIBridge()
            logger.info("initialize: checking dependencies")
            let deps = try await cli!.checkDependencies()
            if deps.contains(where: { !$0.value.installed && $0.value.optional != true }) {
                logger.warning("initialize: missing required dependencies")
                setupState = .missingDependencies(deps)
                return
            }
            logger.info("initialize: loading config")
            let config = try await cli!.getConfig()
            if !config.onboarded {
                setupState = .needsOnboarding
            } else {
                setupState = .ready
                logger.info("initialize: ready, refreshing workspaces")
                await refreshWorkspaces()
            }
        } catch CLIBridge.CLIError.ccwNotFound {
            logger.error("initialize: ccw not found in app bundle")
            setupState = .error("ccw binary not found in app bundle")
        } catch {
            logger.error("initialize: unexpected error \(error.localizedDescription, privacy: .public)")
            setupState = .needsOnboarding
        }
    }

    public func refreshWorkspaces() async {
        guard let cli = cli else { return }
        let start = Date()
        logger.info("refreshWorkspaces start (mainThread=\(Thread.isMainThread, privacy: .public))")
        isLoading = true
        defer { isLoading = false }

        do {
            workspaces = try await cli.listWorkspaces()
            staleWorkspaces = try await cli.staleWorkspaces()
            error = nil
            let elapsed = Date().timeIntervalSince(start)
            logger.info("refreshWorkspaces success count=\(workspaces.count, privacy: .public) stale=\(staleWorkspaces.count, privacy: .public) elapsed=\(elapsed, privacy: .public)s")
        } catch {
            self.error = error
            let elapsed = Date().timeIntervalSince(start)
            logger.error("refreshWorkspaces failed elapsed=\(elapsed, privacy: .public)s error=\(error.localizedDescription, privacy: .public)")
        }
    }

    public func openWorkspace(_ id: String, resume: Bool = true) async {
        guard let cli = cli else { return }
        do {
            try await cli.openWorkspace(id, resume: resume)
        } catch {
            self.error = error
        }
    }

    public func createWorkspace(repo: String, branch: String, base: String?) async {
        guard let cli = cli else { return }
        do {
            try await cli.createWorkspace(repo: repo, branch: branch, base: base)
            await refreshWorkspaces()
        } catch {
            self.error = error
        }
    }

    public func removeWorkspace(_ id: String, force: Bool) async {
        guard let cli = cli else { return }
        do {
            try await cli.removeWorkspace(id, force: force)
            await refreshWorkspaces()
        } catch {
            self.error = error
        }
    }

    public func loadWorkspaceInfo(_ id: String) async {
        guard let cli = cli else { return }
        do {
            workspaceInfo = try await cli.workspaceInfo(id)
            showingWorkspaceInfo = true
        } catch {
            self.error = error
        }
    }

    public func listRepos() async -> [String] {
        guard let cli = cli else { return [] }
        do {
            return try await cli.listRepos()
        } catch {
            self.error = error
            return []
        }
    }

    public func loadConfig() async {
        guard let cli = cli else { return }
        do {
            config = try await cli.getConfig()
        } catch {
            self.error = error
        }
    }

    public func setConfig(key: String, value: String) async {
        guard let cli = cli else { return }
        do {
            try await cli.setConfig(key: key, value: value)
        } catch {
            self.error = error
        }
    }

    public func checkDependencies() async -> [String: DepStatus] {
        guard let cli = cli else { return [:] }
        do {
            return try await cli.checkDependencies()
        } catch {
            self.error = error
            return [:]
        }
    }
}
