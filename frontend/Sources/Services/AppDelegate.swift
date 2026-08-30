import AppKit
import Foundation

final class AppDelegate: NSObject, NSApplicationDelegate {
    private var process: Process?

    func applicationDidFinishLaunching(_ notification: Notification) {
        AppDelegate.shortenToolTipDelay()
        presentAsForegroundApplication()
        startBundledBackendIfNeeded()
    }

    /// The system waits a long time before showing a tooltip, which is a poor
    /// fit for a toolbar of icons where the tooltip is how a button explains
    /// itself. Registered as a default rather than set outright, so anyone who
    /// has chosen their own delay keeps it.
    static func shortenToolTipDelay(in defaults: UserDefaults = .standard) {
        defaults.register(defaults: [toolTipDelayKey: toolTipDelayMilliseconds])
    }

    /// Long enough not to flicker while the pointer crosses the toolbar, short
    /// enough to answer a deliberate hover.
    static let toolTipDelayMilliseconds = 300

    /// AppKit reads the delay, in milliseconds, from this default.
    static let toolTipDelayKey = "NSInitialToolTipDelay"

    func applicationWillTerminate(_ notification: Notification) {
        stopBundledBackend()
    }

    /// A SwiftPM executable is not an application bundle, so AppKit launches it
    /// with the `.prohibited` activation policy. Such a process has no Dock icon
    /// and never owns the menu bar, which also leaves the standard Edit
    /// shortcuts such as Command-V and Command-A without any effect inside text
    /// fields. The packaged application already launches as a regular
    /// application and is left untouched.
    private func presentAsForegroundApplication() {
        guard NSApp.activationPolicy() != .regular else { return }

        NSApp.setActivationPolicy(.regular)
        if let iconURL = AppDelegate.repositoryIconURL(startingAt: Bundle.main.bundleURL),
           let icon = NSImage(contentsOf: iconURL) {
            NSApp.applicationIconImage = icon
        }
        NSApp.activate()
    }

    /// Looks for the repository icon by walking up from the executable, so an
    /// unbundled run shows the real icon instead of a generic placeholder.
    static func repositoryIconURL(startingAt directory: URL) -> URL? {
        var current = directory
        for _ in 0..<8 {
            let candidate = current.appendingPathComponent("build/darwin/icons.icns")
            if FileManager.default.fileExists(atPath: candidate.path) {
                return candidate
            }

            let parent = current.deletingLastPathComponent()
            guard parent.path != current.path else { return nil }
            current = parent
        }
        return nil
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
