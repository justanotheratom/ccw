import Foundation
import SwiftUI

@MainActor
public final class AppState: ObservableObject {
    @Published public var workspaces: [WorkspaceStatus] = []
    @Published public var isLoading = false
    @Published public var error: Error?
    @Published public var workspaceInfo: WorkspaceInfo?
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
        Task { await initialize() }
    }

    private func initialize() async {
        do {
            cli = try CLIBridge()
            let deps = try await cli!.checkDependencies()
            if deps.contains(where: { !$0.value.installed && $0.value.optional != true }) {
                setupState = .missingDependencies(deps)
                return
            }
            let config = try await cli!.getConfig()
            if !config.onboarded {
                setupState = .needsOnboarding
            } else {
                setupState = .ready
                await refreshWorkspaces()
            }
        } catch CLIBridge.CLIError.ccwNotFound {
            setupState = .error("ccw binary not found in app bundle")
        } catch {
            setupState = .needsOnboarding
        }
    }

    public func refreshWorkspaces() async {
        guard let cli = cli else { return }
        isLoading = true
        defer { isLoading = false }

        do {
            workspaces = try await cli.listWorkspaces()
            error = nil
        } catch {
            self.error = error
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
        } catch {
            self.error = error
        }
    }
}
