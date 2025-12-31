import SwiftUI
import Dispatch
import CCWMenubarFeature
import KeyboardShortcuts

@main
struct CCWMenubarApp: App {
    @StateObject private var appState: AppState
    @StateObject private var menuState: MenuState
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    private let logger = CCWLog.ui

    init() {
        let appState = AppState()
        _appState = StateObject(wrappedValue: appState)
        let state = MenuState()
        _menuState = StateObject(wrappedValue: state)
        KeyboardShortcuts.onKeyUp(for: .toggleMenu) {
            DispatchQueue.main.async {
                state.isInserted.toggle()
            }
        }
        DispatchQueue.main.async {
            appState.start()
        }
    }

    var body: some Scene {
        MenuBarExtra("CCW", systemImage: statusImageName, isInserted: isInsertedBinding) {
            MenuBarView()
                .environmentObject(appState)
        }
        .menuBarExtraStyle(.menu)
        .onChange(of: menuState.isInserted) { newValue in
            logger.info("menu bar extra isInserted=\(newValue, privacy: .public)")
            NSLog("CCWMenubar[ui] menu bar extra isInserted=\(newValue)")
            if newValue {
                Task { await appState.refreshWorkspaces() }
            }
        }

        Settings {
            SettingsView()
                .environmentObject(appState)
        }

    }

    private var statusImageName: String {
        if appState.workspaces.contains(where: { $0.state == .connected }) {
            return "circle.fill"
        }
        if appState.workspaces.contains(where: { $0.state == .alive }) {
            return "circle"
        }
        if appState.workspaces.contains(where: { $0.state == .dead }) {
            return "xmark.circle"
        }
        return "terminal.fill"
    }

    private var isInsertedBinding: Binding<Bool> {
        Binding(
            get: { menuState.isInserted },
            set: { newValue in
                if menuState.isInserted != newValue {
                    menuState.isInserted = newValue
                }
            }
        )
    }
}

final class MenuState: ObservableObject {
    @Published var isInserted = true {
        didSet {
            NSLog("CCWMenubar[ui] menuState isInserted didSet value=\(isInserted)")
        }
    }
}


final class AppDelegate: NSObject, NSApplicationDelegate {
    private var signalSources: [DispatchSourceSignal] = []
    private var keepAliveWindow: NSWindow?
    private var keepAliveActivity: NSObjectProtocol?

    func applicationDidFinishLaunching(_ notification: Notification) {
        NSLog("CCWMenubar[delegate] applicationDidFinishLaunching")
        keepAliveActivity = ProcessInfo.processInfo.beginActivity(options: [.automaticTerminationDisabled], reason: "Keep CCW Menubar alive")
        NSLog("CCWMenubar[exit] beginActivity automaticTerminationDisabled")
        createKeepAliveWindow()
        setupTerminationLogging()
    }

    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        NSLog("CCWMenubar[delegate] applicationShouldTerminate")
        return .terminateNow
    }

    func applicationWillTerminate(_ notification: Notification) {
        NSLog("CCWMenubar[delegate] applicationWillTerminate")
        if let activity = keepAliveActivity {
            ProcessInfo.processInfo.endActivity(activity)
            keepAliveActivity = nil
            NSLog("CCWMenubar[exit] endActivity automaticTerminationDisabled")
        }
    }

    private func createKeepAliveWindow() {
        if keepAliveWindow != nil { return }
        let frame = NSRect(x: -10000, y: -10000, width: 1, height: 1)
        let window = NSWindow(contentRect: frame, styleMask: [.borderless], backing: .buffered, defer: false)
        window.isReleasedWhenClosed = false
        window.isOpaque = false
        window.backgroundColor = .clear
        window.hasShadow = false
        window.ignoresMouseEvents = true
        window.level = .statusBar
        window.collectionBehavior = [.stationary, .ignoresCycle, .fullScreenAuxiliary]
        window.orderFront(nil)
        keepAliveWindow = window
        NSLog("CCWMenubar[exit] keep-alive window created")
    }

    private func setupTerminationLogging() {
        atexit {
            NSLog("CCWMenubar[exit] atexit invoked")
        }
        let signals: [Int32] = [SIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGABRT]
        for sig in signals {
            signal(sig, SIG_IGN)
            let source = DispatchSource.makeSignalSource(signal: sig, queue: .main)
            source.setEventHandler {
                NSLog("CCWMenubar[exit] signal \(sig)")
            }
            source.resume()
            signalSources.append(source)
        }
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        NSLog("CCWMenubar[delegate] applicationShouldTerminateAfterLastWindowClosed")
        return false
    }
}
