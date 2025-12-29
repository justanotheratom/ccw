import SwiftUI

struct StatusIndicator: View {
    let state: WorkspaceState

    var body: some View {
        Circle()
            .fill(color)
            .frame(width: 8, height: 8)
    }

    private var color: Color {
        switch state {
        case .connected:
            return .green
        case .alive:
            return .yellow
        case .dead:
            return .red
        }
    }
}
