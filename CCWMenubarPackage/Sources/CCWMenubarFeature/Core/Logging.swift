import Foundation
import os

public enum CCWLog {
    public static let subsystem = "com.justanotheratom.ccw-menubar"
    public static let appState = Logger(subsystem: subsystem, category: "app-state")
    public static let cli = Logger(subsystem: subsystem, category: "cli")
    public static let ui = Logger(subsystem: subsystem, category: "ui")
}
