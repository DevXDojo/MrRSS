import SwiftUI

struct ServerSettingsView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var didSave = false

    var body: some View {
        Form {
            Section(t("client.connection.backendServer")) {
                TextField(t("client.connection.serverAddress"), text: $viewModel.serverURLText)
                    .textFieldStyle(.roundedBorder)

                Text(t("client.connection.serverAddressHelp"))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }

            HStack {
                Circle()
                    .fill(viewModel.connectionState.color)
                    .frame(width: 8, height: 8)
                Text(viewModel.connectionState.title)
                    .foregroundStyle(.secondary)
                Spacer()
                if didSave {
                    Text(t("client.connection.saved"))
                        .foregroundStyle(.secondary)
                }
                Button(t("client.connection.saveAndReconnect")) {
                    didSave = viewModel.saveServerAddress()
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .formStyle(.grouped)
        .padding(8)
    }
}
