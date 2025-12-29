import SwiftUI

public struct MenuBarView: View {
    @EnvironmentObject private var appState: AppState

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
    }

    private var header: some View {
        HStack {
            Text("CCW Workspaces")
                .font(.headline)
            Spacer()
            Button(action: {}) {
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
            Text("Missing dependencies")
                .padding()
        case .needsOnboarding:
            Text("Complete setup in Settings")
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
