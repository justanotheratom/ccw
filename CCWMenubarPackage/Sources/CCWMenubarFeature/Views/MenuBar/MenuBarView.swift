import SwiftUI

// Navigation destinations
enum MenuDestination: Hashable {
    case settings
    case newWorkspace
    case onboarding
    case workspaceInfo(WorkspaceInfo)
}

private enum AutoTerminationGuard {
    static var disabled = false
}

public struct MenuBarView: View {
    @EnvironmentObject private var appState: AppState
    @State private var navigationPath = NavigationPath()

    public init() {}

    public var body: some View {
        NavigationStack(path: $navigationPath) {
            mainContent
                .navigationDestination(for: MenuDestination.self) { destination in
                    switch destination {
                    case .settings:
                        SettingsView(navigationPath: $navigationPath)
                            .environmentObject(appState)
                    case .newWorkspace:
                        NewWorkspaceView(navigationPath: $navigationPath)
                            .environmentObject(appState)
                    case .onboarding:
                        OnboardingView()
                            .environmentObject(appState)
                    case .workspaceInfo(let info):
                        WorkspaceInfoView(info: info)
                    }
                }
        }
        .frame(width: 320)
        .onAppear {
            if !AutoTerminationGuard.disabled {
                AutoTerminationGuard.disabled = true
                ProcessInfo.processInfo.disableAutomaticTermination("Keep CCW Menubar alive")
                NSLog("CCWMenubar[exit] automatic termination disabled (menu bar view)")
            }
            NSLog("CCWMenubar[ui] menu bar view onAppear")
            Task { await appState.refreshWorkspaces() }
        }
        .onDisappear {
            NSLog("CCWMenubar[ui] menu bar view onDisappear")
        }
        .onChange(of: appState.showingWorkspaceInfo) { _, showing in
            if showing, let info = appState.workspaceInfo {
                navigationPath.append(MenuDestination.workspaceInfo(info))
                appState.showingWorkspaceInfo = false
            }
        }
    }

    private var mainContent: some View {
        VStack(alignment: .leading, spacing: 0) {
            header

            Divider()
                .opacity(0.3)

            content

            Divider()
                .opacity(0.3)

            footer
        }
        .background(.ultraThinMaterial)
    }

    private var header: some View {
        HStack {
            Text("WORKSPACES")
                .font(.system(size: 11, weight: .medium))
                .foregroundStyle(.secondary)
                .tracking(0.5)

            Spacer()

            Button(action: {
                navigationPath.append(MenuDestination.newWorkspace)
            }) {
                Image(systemName: "plus")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(.primary)
            }
            .buttonStyle(.plain)
            .contentShape(Rectangle())
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
    }

    @ViewBuilder
    private var content: some View {
        switch appState.setupState {
        case .checking:
            HStack {
                Spacer()
                ProgressView()
                    .controlSize(.small)
                Spacer()
            }
            .padding(.vertical, 20)

        case .missingDependencies:
            VStack(alignment: .leading, spacing: 8) {
                Label("Missing dependencies", systemImage: "exclamationmark.triangle")
                    .foregroundStyle(.orange)
                    .font(.system(size: 13))

                Button("Open Setup") {
                    navigationPath.append(MenuDestination.onboarding)
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
            }
            .padding(14)

        case .needsOnboarding:
            VStack(alignment: .leading, spacing: 8) {
                Label("Setup required", systemImage: "gear.badge")
                    .font(.system(size: 13))

                Button("Open Setup") {
                    navigationPath.append(MenuDestination.onboarding)
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
            }
            .padding(14)

        case .error(let message):
            VStack(alignment: .leading, spacing: 4) {
                Label("Error", systemImage: "xmark.circle")
                    .foregroundStyle(.red)
                    .font(.system(size: 13, weight: .medium))
                Text(message)
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
            .padding(14)

        case .ready:
            if appState.isLoading {
                HStack {
                    Spacer()
                    ProgressView()
                        .controlSize(.small)
                    Spacer()
                }
                .padding(.vertical, 20)
            } else if appState.workspaces.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "folder.badge.questionmark")
                        .font(.system(size: 24))
                        .foregroundStyle(.tertiary)
                    Text("No workspaces")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 20)
            } else {
                VStack(spacing: 0) {
                    ForEach(appState.workspaces) { workspace in
                        WorkspaceRow(workspace: workspace)
                    }
                }

                if !appState.staleWorkspaces.isEmpty {
                    Divider()
                        .opacity(0.3)

                    Menu {
                        ForEach(appState.staleWorkspaces) { workspace in
                            Button(workspace.id) {
                                Task { await appState.openWorkspace(workspace.id) }
                            }
                        }
                    } label: {
                        HStack {
                            Label("\(appState.staleWorkspaces.count) Stale", systemImage: "clock.badge.exclamationmark")
                                .font(.system(size: 13))
                                .foregroundStyle(.secondary)
                            Spacer()
                            Image(systemName: "chevron.right")
                                .font(.system(size: 10, weight: .semibold))
                                .foregroundStyle(.tertiary)
                        }
                        .padding(.horizontal, 14)
                        .padding(.vertical, 10)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                }
            }
        }
    }

    private var footer: some View {
        VStack(alignment: .leading, spacing: 0) {
            Button(action: {
                navigationPath.append(MenuDestination.settings)
            }) {
                HStack {
                    Label("Settings", systemImage: "gear")
                        .font(.system(size: 13))
                    Spacer()
                    Text("⌘,")
                        .font(.system(size: 11))
                        .foregroundStyle(.tertiary)
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .keyboardShortcut(",", modifiers: .command)

            Button(action: {
                NSLog("CCWMenubar[ui] quit tapped")
                NSApp.terminate(nil)
            }) {
                HStack {
                    Label("Quit", systemImage: "power")
                        .font(.system(size: 13))
                    Spacer()
                    Text("⌘Q")
                        .font(.system(size: 11))
                        .foregroundStyle(.tertiary)
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .keyboardShortcut("q", modifiers: .command)
        }
    }
}
