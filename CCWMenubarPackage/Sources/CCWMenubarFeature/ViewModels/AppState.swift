import Foundation
import SwiftUI

@MainActor
public final class AppState: ObservableObject {
    private let logger = CCWLog.appState
    @Published public var workspaces: [WorkspaceStatus] = [] {
        didSet { NSLog("CCWMenubar[app-state] workspaces didSet count=\(workspaces.count)") }
    }
    @Published public var staleWorkspaces: [WorkspaceStatus] = [] {
        didSet { NSLog("CCWMenubar[app-state] staleWorkspaces didSet count=\(staleWorkspaces.count)") }
    }
    @Published public var isLoading = false {
        didSet { NSLog("CCWMenubar[app-state] isLoading didSet value=\(isLoading)") }
    }
    @Published public var error: Error?
    @Published public var workspaceInfo: WorkspaceInfo?
    @Published public var showingWorkspaceInfo = false
    @Published public var config: CCWConfig?
    @Published public var setupState: SetupState = .checking {
        didSet { NSLog("CCWMenubar[app-state] setupState didSet value=\(String(describing: setupState))") }
    }

    public enum SetupState {
        case checking
        case needsOnboarding
        case missingDependencies([String: DepStatus])
        case ready
        case error(String)
    }

    private var cli: CLIBridge?
    private var didStart = false

    public init() {
        logger.info("init start (mainThread=\(Thread.isMainThread, privacy: .public))")
        NSLog("CCWMenubar[app-state] init start (mainThread=\(Thread.isMainThread))")
    }

    public func start() {
        guard !didStart else { return }
        didStart = true
        Task {
            await Task.yield()
            await initialize()
        }
    }

    private func initialize() async {
        await Task.yield()
        do {
            logger.info("initialize: creating CLI bridge")
            NSLog("CCWMenubar[app-state] initialize: creating CLI bridge")
            cli = try CLIBridge()
            logger.info("initialize: checking dependencies")
            NSLog("CCWMenubar[app-state] initialize: checking dependencies")
            let deps = try await cli!.checkDependencies()
            if deps.contains(where: { !$0.value.installed && $0.value.optional != true }) {
                logger.warning("initialize: missing required dependencies")
                NSLog("CCWMenubar[app-state] initialize: missing required dependencies")
                setupState = .missingDependencies(deps)
                return
            }
            logger.info("initialize: loading config")
            NSLog("CCWMenubar[app-state] initialize: loading config")
            let config = try await cli!.getConfig()
            if !config.onboarded {
                setupState = .needsOnboarding
            } else {
                setupState = .ready
                logger.info("initialize: ready, refreshing workspaces")
                NSLog("CCWMenubar[app-state] initialize: ready, refreshing workspaces")
                await refreshWorkspaces()
            }
        } catch CLIBridge.CLIError.ccwNotFound {
            logger.error("initialize: ccw not found in app bundle")
            NSLog("CCWMenubar[app-state] initialize: ccw not found in app bundle")
            setupState = .error("ccw binary not found in app bundle")
        } catch {
            logger.error("initialize: unexpected error \(error.localizedDescription, privacy: .public)")
            NSLog("CCWMenubar[app-state] initialize: unexpected error \(error.localizedDescription)")
            setupState = .needsOnboarding
        }
    }

    public func refreshWorkspaces() async {
        await Task.yield()
        guard let cli = cli else { return }
        guard !isLoading else { return }
        let start = Date()
        logger.info("refreshWorkspaces start (mainThread=\(Thread.isMainThread, privacy: .public))")
        NSLog("CCWMenubar[app-state] refreshWorkspaces start (mainThread=\(Thread.isMainThread))")
        isLoading = true
        defer { isLoading = false }

        do {
            workspaces = try await cli.listWorkspaces()
            NSLog("CCWMenubar[app-state] listWorkspaces done count=\(workspaces.count)")
            NSLog("CCWMenubar[app-state] staleWorkspaces start")
            staleWorkspaces = try await cli.staleWorkspaces()
            error = nil
            let elapsed = Date().timeIntervalSince(start)
            logger.info("refreshWorkspaces success count=\(self.workspaces.count, privacy: .public) stale=\(self.staleWorkspaces.count, privacy: .public) elapsed=\(elapsed, privacy: .public)s")
            NSLog("CCWMenubar[app-state] refreshWorkspaces success count=\(self.workspaces.count) stale=\(self.staleWorkspaces.count) elapsed=\(elapsed)s")
        } catch {
            self.error = error
            let elapsed = Date().timeIntervalSince(start)
            logger.error("refreshWorkspaces failed elapsed=\(elapsed, privacy: .public)s error=\(error.localizedDescription, privacy: .public)")
            NSLog("CCWMenubar[app-state] refreshWorkspaces failed elapsed=\(elapsed)s error=\(error.localizedDescription)")
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
