import SwiftUI

@main
struct MrRSSApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var viewModel = AppViewModel(autoLoad: false)
    @StateObject private var localization = Localization.shared

    var body: some Scene {
        WindowGroup {
            ContentView(viewModel: viewModel)
                .preferredColorScheme(viewModel.preferredColorScheme)
                // Interface strings are read through `t(_:)` rather than through
                // the environment, so the tree is rebuilt when the language changes.
                .id(localization.language)
                .environmentObject(localization)
                .task {
                    await viewModel.start()
                }
        }
        .defaultSize(width: 1280, height: 780)
        .commands {
            SidebarCommands()
            CommandGroup(after: .newItem) {
                Button("Refresh") {
                    viewModel.refreshFromSources()
                }
                .keyboardShortcut("r", modifiers: .command)
            }
        }

        Settings {
            SettingsRootView(viewModel: viewModel)
        }
    }
}
