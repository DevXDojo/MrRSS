import SwiftUI

@main
struct MrRSSApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var viewModel = AppViewModel(autoLoad: false)

    var body: some Scene {
        WindowGroup {
            ContentView(viewModel: viewModel)
                .preferredColorScheme(viewModel.preferredColorScheme)
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
