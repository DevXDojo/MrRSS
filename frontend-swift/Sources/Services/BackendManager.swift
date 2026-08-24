import AppKit
import Foundation

final class BackendManager: NSObject, NSApplicationDelegate {
    private var process: Process?

    func applicationDidFinishLaunching(_ notification: Notification) {
        startBundledBackendIfNeeded()
    }

    func applicationWillTerminate(_ notification: Notification) {
        stopBundledBackend()
    }

    private func startBundledBackendIfNeeded() {
        let baseURL = ServerConfiguration.savedBaseURL
        guard ["127.0.0.1", "localhost"].contains(baseURL.host?.lowercased() ?? ""),
              baseURL.port ?? 80 == 1234,
              let executableURL = Bundle.main.url(forResource: "mrrss-server", withExtension: nil) else {
            return
        }

        let supportDirectory = FileManager.default.urls(
            for: .applicationSupportDirectory,
            in: .userDomainMask
        )[0].appendingPathComponent("MrRSS-SwiftUI", isDirectory: true)

        do {
            try FileManager.default.createDirectory(
                at: supportDirectory.appendingPathComponent("data", isDirectory: true),
                withIntermediateDirectories: true
            )
            let process = Process()
            process.executableURL = executableURL
            process.arguments = ["-host", "127.0.0.1", "-port", "1234"]
            process.currentDirectoryURL = supportDirectory
            process.standardOutput = FileHandle.nullDevice
            process.standardError = FileHandle.nullDevice
            try process.run()
            self.process = process
        } catch {
            NSLog("Unable to start bundled MrRSS backend: %@", error.localizedDescription)
        }
    }

    private func stopBundledBackend() {
        guard let process, process.isRunning else { return }
        process.terminate()
        process.waitUntilExit()
        self.process = nil
    }
}
