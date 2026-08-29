import AppKit
import SwiftUI

/// Restores the window position and size the backend remembers, and saves it
/// again when the window moves, resizes or closes, as the previous interface did.
struct WindowStatePersistence: ViewModifier {
    @ObservedObject var viewModel: AppViewModel
    @State private var observers: [NSObjectProtocol] = []
    @State private var didRestore = false

    func body(content: Content) -> some View {
        content
            .background(WindowReader { window in
                guard !didRestore else { return }
                didRestore = true
                Task { await restore(window) }
                observe(window)
            })
    }

    private func restore(_ window: NSWindow) async {
        guard let state = try? await viewModel.api.fetchWindowState() else { return }
        guard state.width > 200, state.height > 200 else { return }

        var frame = NSRect(
            x: CGFloat(state.x),
            y: CGFloat(state.y),
            width: CGFloat(state.width),
            height: CGFloat(state.height)
        )

        // A saved position from a display that is no longer attached would put
        // the window out of reach, so fall back to the centre of the screen.
        let visible = NSScreen.screens.map(\.visibleFrame)
        if !visible.contains(where: { $0.intersects(frame) }) {
            frame.origin = CGPoint(
                x: (NSScreen.main?.visibleFrame.midX ?? 0) - frame.width / 2,
                y: (NSScreen.main?.visibleFrame.midY ?? 0) - frame.height / 2
            )
        }

        window.setFrame(frame, display: true)
        if state.maximized {
            window.zoom(nil)
        }
    }

    private func observe(_ window: NSWindow) {
        let center = NotificationCenter.default
        let names: [Notification.Name] = [
            NSWindow.didMoveNotification,
            NSWindow.didResizeNotification,
            NSWindow.willCloseNotification
        ]
        observers = names.map { name in
            center.addObserver(forName: name, object: window, queue: .main) { _ in
                MainActor.assumeIsolated { save(window) }
            }
        }
    }

    private func save(_ window: NSWindow) {
        let frame = window.frame
        let state = WindowState(
            x: Int(frame.origin.x),
            y: Int(frame.origin.y),
            width: Int(frame.width),
            height: Int(frame.height),
            maximized: window.isZoomed
        )
        Task { try? await viewModel.api.saveWindowState(state) }
    }
}

/// Hands the enclosing window to a closure once it exists.
private struct WindowReader: NSViewRepresentable {
    let onWindow: (NSWindow) -> Void

    func makeNSView(context: Context) -> NSView {
        let view = NSView()
        DispatchQueue.main.async {
            if let window = view.window {
                onWindow(window)
            }
        }
        return view
    }

    func updateNSView(_ nsView: NSView, context: Context) {
        guard let window = nsView.window else { return }
        onWindow(window)
    }
}

extension View {
    /// Remembers where the window was between launches.
    func persistingWindowState(with viewModel: AppViewModel) -> some View {
        modifier(WindowStatePersistence(viewModel: viewModel))
    }
}
