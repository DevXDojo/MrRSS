import SwiftUI

struct ServerSettingsView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var didSave = false

    var body: some View {
        Form {
            Section("Backend server") {
                TextField("API address", text: $viewModel.serverURLText)
                    .textFieldStyle(.roundedBorder)

                Text("The default server address is http://127.0.0.1:1234/api. You can also set MRRSS_API_BASE_URL before launching the app.")
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
                    Text("Saved")
                        .foregroundStyle(.secondary)
                }
                Button("Save and reconnect") {
                    didSave = viewModel.saveServerAddress()
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .formStyle(.grouped)
        .padding(8)
    }
}
