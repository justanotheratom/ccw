import SwiftUI

public struct MenuBarView: View {
    @EnvironmentObject private var appState: AppState
    @State private var showingNewWorkspace = false
    @State private var showingOnboarding = false

    public init() {}

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
            Divider()

            content

            Divider()
            footer
        }
        .frame(width: 350)
        .onAppear {
            Task { await appState.refreshWorkspaces() }
        }
        .sheet(isPresented: $showingNewWorkspace) {
            NewWorkspaceView()
                .environmentObject(appState)
        }
        .sheet(isPresented: $showingOnboarding) {
            OnboardingView()
                .environmentObject(appState)
        }
        .sheet(isPresented: $appState.showingWorkspaceInfo) {
            WorkspaceInfoView(info: appState.workspaceInfo)
        }
    }

    private var header: some View {
        HStack {
            Text("CCW Workspaces")
                .font(.headline)
            Spacer()
            Button(action: { showingNewWorkspace = true }) {
                Image(systemName: "plus")
            }
        }
        .padding()
    }

    @ViewBuilder
    private var content: some View {
        switch appState.setupState {
        case .checking:
            ProgressView()
                .padding()
        case .missingDependencies:
            VStack(alignment: .leading, spacing: 8) {
                Text("Missing dependencies")
                Button("Open Setup") { showingOnboarding = true }
            }
            .padding()
        case .needsOnboarding:
            VStack(alignment: .leading, spacing: 8) {
                Text("Setup required")
                Button("Open Setup") { showingOnboarding = true }
            }
            .padding()
        case .error(let message):
            Text(message)
                .padding()
        case .ready:
            if appState.isLoading {
                ProgressView()
                    .padding()
            } else if appState.workspaces.isEmpty {
                Text("No workspaces found")
                    .padding()
            } else {
                ForEach(appState.workspaces) { workspace in
                    WorkspaceRow(workspace: workspace)
                }
                if !appState.staleWorkspaces.isEmpty {
                    Divider()
                    Menu("Stale Workspaces (\(appState.staleWorkspaces.count))") {
                        ForEach(appState.staleWorkspaces) { workspace in
                            Button(workspace.id) {
                                Task { await appState.openWorkspace(workspace.id) }
                            }
                        }
                    }
                    .padding(.horizontal)
                    .padding(.vertical, 6)
                }
            }
        }
    }

    private var footer: some View {
        VStack(alignment: .leading, spacing: 8) {
            Button("Settings...") {
                NSApp.sendAction(Selector(("showSettingsWindow:")), to: nil, from: nil)
            }
            .keyboardShortcut(",", modifiers: .command)

            Button("Quit") {
                NSApp.terminate(nil)
            }
            .keyboardShortcut("q", modifiers: .command)
        }
        .padding()
    }
}
