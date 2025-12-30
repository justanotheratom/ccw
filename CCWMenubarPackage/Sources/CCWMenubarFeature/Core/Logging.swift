import Foundation
import os

enum CCWLog {
    static let subsystem = "com.justanotheratom.ccw-menubar"
    static let appState = Logger(subsystem: subsystem, category: "app-state")
    static let cli = Logger(subsystem: subsystem, category: "cli")
    static let ui = Logger(subsystem: subsystem, category: "ui")
}
