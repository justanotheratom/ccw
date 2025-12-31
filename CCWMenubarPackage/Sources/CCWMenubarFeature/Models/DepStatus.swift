import Foundation

public struct DepStatus: Codable, Sendable {
    public let installed: Bool
    public let path: String
    public let optional: Bool?
}
