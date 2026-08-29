import SwiftUI

/// Shows the short confirmations the view model reports, in place of the toast
/// the previous interface used.
struct StatusOverlay: View {
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        VStack {
            Spacer()
            if let message = viewModel.statusMessage {
                Text(message)
                    .font(.callout)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 10)
                    .background(.regularMaterial, in: Capsule())
                    .shadow(radius: 8, y: 2)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
                    .task(id: message) {
                        // Long enough to read, short enough to stay out of the way.
                        try? await Task.sleep(for: .seconds(3))
                        withAnimation { viewModel.clearStatusMessage() }
                    }
            }
        }
        .padding(.bottom, 24)
        .animation(.easeInOut(duration: 0.2), value: viewModel.statusMessage)
        .allowsHitTesting(false)
    }
}
