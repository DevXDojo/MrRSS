import SwiftUI

/// Adds a subscription or edits an existing one. The form covers every option
/// the previous interface offered: the source, its category and tags, the
/// refresh and proxy overrides, XPath scraping and newsletter collection.
struct FeedEditorView: View {
    enum Mode: Equatable {
        case add
        case edit(Feed)

        var isEditing: Bool {
            if case .edit = self { return true }
            return false
        }
    }

    /// Which kind of source the form is describing.
    enum SourceKind: String, CaseIterable, Identifiable {
        case regular
        case xpath
        case email
        case script

        var id: String { rawValue }

        var title: String {
            switch self {
            case .regular: t("modal.feed.typeRegular")
            case .xpath: t("modal.feed.typeXPath")
            case .email: t("modal.feed.typeEmail")
            case .script: t("modal.feed.typeCustomScript")
            }
        }
    }

    let mode: Mode
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var draft = FeedDraft()
    @State private var sourceKind: SourceKind = .regular
    @State private var showsAdvanced = false
    @State private var isSubmitting = false
    @State private var imapMessage: String?
    @State private var isTestingIMAP = false

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()

            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    sourceSection
                    organisationSection

                    DisclosureTrigger(
                        isExpanded: $showsAdvanced,
                        collapsedTitle: t("setting.reading.showAdvancedSettings"),
                        expandedTitle: t("setting.reading.hideAdvancedSettings")
                    )

                    if showsAdvanced {
                        refreshSection
                        proxySection
                        readingSection
                    }
                }
                .padding(22)
            }

            Divider()
            footer
        }
        .frame(width: 620, height: 640)
        .onAppear(perform: loadDraft)
    }

    // MARK: - Sections

    private var header: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(mode.isEditing ? t("modal.feed.editFeed") : t("modal.feed.addNewFeed"))
                .font(.title2.bold())
            Text(t("modal.feed.rssUrl"))
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(22)
    }

    private var sourceSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Picker(t("modal.feed.feedDiscovery"), selection: $sourceKind) {
                ForEach(SourceKind.allCases) { kind in
                    Text(kind.title).tag(kind)
                }
            }
            .pickerStyle(.segmented)
            .labelsHidden()

            switch sourceKind {
            case .regular:
                labelledField(t("modal.feed.rssUrl"), text: $draft.url, prompt: "https://example.com/feed.xml")
            case .xpath:
                labelledField(t("modal.feed.sourceUrl"), text: $draft.url, prompt: t("modal.feed.sourceUrlPlaceholder"))
                xpathFields
            case .email:
                emailFields
            case .script:
                labelledField(t("modal.feed.rssUrl"), text: $draft.url, prompt: "https://example.com/feed.xml")
                labelledField(t("setting.customization.script"), text: $draft.scriptPath, prompt: "script.js")
            }

            labelledField(t("common.form.title"), text: $draft.title, prompt: t("modal.feed.titlePlaceholder"))
        }
    }

    private var xpathFields: some View {
        VStack(alignment: .leading, spacing: 10) {
            Picker(t("modal.feed.xpathType"), selection: $draft.type) {
                Text(t("modal.feed.xpathHtml")).tag("HTML+XPath")
                Text(t("modal.feed.xpathXml")).tag("XML+XPath")
            }

            labelledField(t("modal.feed.xpathItem"), text: $draft.xPathItem, prompt: "//article")
            Text(t("modal.feed.xpathItemHelp"))
                .font(.caption)
                .foregroundStyle(.secondary)

            labelledField(t("modal.feed.xpathItemTitle"), text: $draft.xPathItemTitle, prompt: ".//h2")
            labelledField(t("modal.feed.xpathItemContent"), text: $draft.xPathItemContent, prompt: ".//p")
            labelledField(t("modal.feed.xpathItemUri"), text: $draft.xPathItemURI, prompt: ".//a/@href")
            labelledField(t("modal.feed.xpathItemAuthor"), text: $draft.xPathItemAuthor, prompt: "")
            labelledField(t("modal.feed.xpathItemTimestamp"), text: $draft.xPathItemTimestamp, prompt: "")
            labelledField(t("modal.feed.xpathItemTimeFormat"), text: $draft.xPathItemTimeFormat, prompt: "")
            labelledField(t("modal.feed.xpathItemThumbnail"), text: $draft.xPathItemThumbnail, prompt: "")
            labelledField(t("modal.feed.xpathItemCategories"), text: $draft.xPathItemCategories, prompt: "")
            labelledField(t("modal.feed.xpathItemUid"), text: $draft.xPathItemUID, prompt: "")
        }
    }

    private var emailFields: some View {
        VStack(alignment: .leading, spacing: 10) {
            labelledField(t("modal.feed.emailServer"), text: $draft.emailIMAPServer, prompt: "imap.example.com")

            HStack(alignment: .bottom, spacing: 10) {
                VStack(alignment: .leading, spacing: 5) {
                    Text(t("modal.feed.emailFolder"))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    TextField("", text: $draft.emailFolder, prompt: Text("INBOX"))
                        .textFieldStyle(.roundedBorder)
                }
                VStack(alignment: .leading, spacing: 5) {
                    Text(t("client.feed.imapPort"))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    TextField("", value: $draft.emailIMAPPort, format: .number)
                        .textFieldStyle(.roundedBorder)
                        .frame(width: 90)
                }
            }

            labelledField(t("modal.feed.emailUsername"), text: $draft.emailUsername, prompt: "")

            VStack(alignment: .leading, spacing: 5) {
                Text(t("modal.feed.emailPassword"))
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                SecureField("", text: $draft.emailPassword, prompt: Text(t("modal.feed.emailPasswordPlaceholder")))
                    .textFieldStyle(.roundedBorder)
            }

            labelledField(t("modal.feed.emailAddress"), text: $draft.emailAddress, prompt: "")
            Text(t("modal.feed.emailAddressHint"))
                .font(.caption)
                .foregroundStyle(.secondary)

            HStack(spacing: 10) {
                Button(t("modal.feed.emailTestConnection")) {
                    Task { await testIMAP() }
                }
                .disabled(isTestingIMAP)

                if isTestingIMAP {
                    ProgressView().controlSize(.small)
                }
                if let imapMessage {
                    Text(imapMessage)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
    }

    private var organisationSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            VStack(alignment: .leading, spacing: 5) {
                Text(t("modal.feed.feedCategory"))
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                HStack(spacing: 8) {
                    TextField("", text: $draft.category, prompt: Text(t("modal.feed.categoryPlaceholder")))
                        .textFieldStyle(.roundedBorder)
                    if !viewModel.folders.isEmpty {
                        Menu {
                            Button(t("sidebar.feedList.uncategorized")) { draft.category = "" }
                            ForEach(viewModel.folders, id: \.self) { folder in
                                Button(folder) { draft.category = folder }
                            }
                        } label: {
                            Image(systemName: "folder")
                        }
                        .menuStyle(.borderlessButton)
                        .frame(width: 30)
                    }
                }
            }

            if !viewModel.tags.isEmpty {
                VStack(alignment: .leading, spacing: 6) {
                    Text(t("modal.feed.feedTags"))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    TagPicker(tags: viewModel.tags, selection: $draft.tags)
                }
            }
        }
    }

    private var refreshSection: some View {
        GroupBox(t("modal.feed.refreshSettings")) {
            VStack(alignment: .leading, spacing: 10) {
                Picker(t("modal.feed.refreshMode"), selection: refreshModeBinding) {
                    Text(t("setting.feed.useGlobalRefresh")).tag(0)
                    Text(t("setting.feed.useIntelligentInterval")).tag(-1)
                    Text(t("setting.feed.neverRefresh")).tag(-2)
                    Text(t("setting.feed.useCustomInterval")).tag(1)
                }

                if draft.refreshInterval > 0 {
                    HStack {
                        Text(t("modal.feed.refreshInterval"))
                        TextField(
                            t("modal.feed.refreshIntervalPlaceholder"),
                            value: $draft.refreshInterval,
                            format: .number
                        )
                        .textFieldStyle(.roundedBorder)
                        .frame(width: 90)
                    }
                }

                Text(t("modal.feed.refreshIntervalDesc"))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            .padding(8)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var proxySection: some View {
        GroupBox(t("modal.feed.proxy")) {
            VStack(alignment: .leading, spacing: 10) {
                Toggle(t("setting.network.enableProxy"), isOn: $draft.proxyEnabled)
                if draft.proxyEnabled {
                    TextField("", text: $draft.proxyURL, prompt: Text("http://127.0.0.1:7890"))
                        .textFieldStyle(.roundedBorder)
                }
                Text(t("modal.feed.proxyDesc"))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            .padding(8)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var readingSection: some View {
        GroupBox(t("setting.tab.readingAndDisplay")) {
            VStack(alignment: .leading, spacing: 10) {
                Toggle(t("setting.reading.hideFromTimeline"), isOn: $draft.hideFromTimeline)
                Text(t("setting.reading.hideFromTimelineDesc"))
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Toggle(t("setting.feed.imageMode"), isOn: $draft.isImageMode)

                Picker(t("setting.feed.articleViewMode"), selection: $draft.articleViewMode) {
                    Text(t("setting.feed.useGlobalSettings")).tag("global")
                    Text(t("setting.reading.viewAsRendered")).tag("rendered")
                    Text(t("setting.reading.viewAsWebpage")).tag("webpage")
                }

                Picker(t("setting.feed.autoExpandContent"), selection: $draft.autoExpandContent) {
                    Text(t("setting.feed.useGlobalSettings")).tag("global")
                    Text(t("common.action.yes")).tag("enabled")
                    Text(t("common.action.no")).tag("disabled")
                }
            }
            .padding(8)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var footer: some View {
        HStack {
            if case .edit(let feed) = mode {
                Button(t("modal.feed.syncFeed")) {
                    Task { await viewModel.refreshFeed(feed) }
                }
            }
            Spacer()
            Button(t("common.cancel"), role: .cancel) { dismiss() }
            Button(mode.isEditing ? t("common.action.saveChanges") : t("modal.feed.addSubscription")) {
                Task { await submit() }
            }
            .keyboardShortcut(.defaultAction)
            .disabled(isSubmitting || !isValid)
        }
        .padding(18)
        .overlay(alignment: .center) {
            if isSubmitting {
                ProgressView().controlSize(.small)
            }
        }
    }

    // MARK: - Behaviour

    private var isValid: Bool {
        switch sourceKind {
        case .email:
            return !draft.emailIMAPServer.trimmingCharacters(in: .whitespaces).isEmpty
                && !draft.emailUsername.trimmingCharacters(in: .whitespaces).isEmpty
        default:
            return !draft.url.trimmingCharacters(in: .whitespaces).isEmpty
        }
    }

    private var refreshModeBinding: Binding<Int> {
        Binding(
            get: {
                switch draft.refreshInterval {
                case 0: 0
                case -1: -1
                case -2: -2
                default: 1
                }
            },
            set: { newValue in
                switch newValue {
                case 0: draft.refreshInterval = 0
                case -1: draft.refreshInterval = -1
                case -2: draft.refreshInterval = -2
                default: draft.refreshInterval = max(1, draft.refreshInterval)
                }
            }
        )
    }

    private func loadDraft() {
        switch mode {
        case .add:
            draft = FeedDraft()
            sourceKind = .regular
        case .edit(let feed):
            draft = FeedDraft(feed: feed, tags: viewModel.feedTags[feed.id] ?? [])
            if feed.isXPathFeed {
                sourceKind = .xpath
            } else if feed.isEmailFeed {
                sourceKind = .email
            } else if !feed.scriptPath.isEmpty {
                sourceKind = .script
            } else {
                sourceKind = .regular
            }
            showsAdvanced = feed.hideFromTimeline || feed.proxyEnabled
                || feed.refreshInterval != 0 || feed.isImageMode
        }
    }

    private func testIMAP() async {
        isTestingIMAP = true
        defer { isTestingIMAP = false }
        do {
            imapMessage = try await viewModel.testIMAPConnection(draft)
        } catch {
            imapMessage = "\(t("modal.feed.errorConnection")): \(error.localizedDescription)"
        }
    }

    private func submit() async {
        isSubmitting = true
        defer { isSubmitting = false }

        var payload = draft
        if sourceKind != .xpath {
            payload.type = ""
        }
        if sourceKind != .script {
            payload.scriptPath = ""
        }
        if sourceKind != .email {
            payload.emailIMAPServer = ""
            payload.emailUsername = ""
            payload.emailPassword = ""
            payload.emailAddress = ""
        }

        if await viewModel.saveFeed(payload, isEditing: mode.isEditing) {
            dismiss()
        }
    }

    private func labelledField(_ label: String, text: Binding<String>, prompt: String) -> some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(label)
                .font(.subheadline)
                .foregroundStyle(.secondary)
            TextField(label, text: text, prompt: Text(prompt))
                .textFieldStyle(.roundedBorder)
                .labelsHidden()
        }
    }
}

/// A row of toggles for assigning tags to a feed.
struct TagPicker: View {
    let tags: [Tag]
    @Binding var selection: [Int]

    private let columns = [GridItem(.adaptive(minimum: 120), spacing: 8)]

    var body: some View {
        LazyVGrid(columns: columns, alignment: .leading, spacing: 8) {
            ForEach(tags) { tag in
                Toggle(isOn: binding(for: tag)) {
                    Label {
                        Text(tag.name).lineLimit(1)
                    } icon: {
                        Circle()
                            .fill(Color(hex: tag.color) ?? .accentColor)
                            .frame(width: 8, height: 8)
                    }
                }
                .toggleStyle(.checkbox)
            }
        }
    }

    private func binding(for tag: Tag) -> Binding<Bool> {
        Binding(
            get: { selection.contains(tag.id) },
            set: { isOn in
                if isOn {
                    if !selection.contains(tag.id) { selection.append(tag.id) }
                } else {
                    selection.removeAll { $0 == tag.id }
                }
            }
        )
    }
}

/// A plain "show or hide the extra options" control.
struct DisclosureTrigger: View {
    @Binding var isExpanded: Bool
    let collapsedTitle: String
    let expandedTitle: String

    var body: some View {
        Button {
            withAnimation(.easeInOut(duration: 0.15)) { isExpanded.toggle() }
        } label: {
            Label(
                isExpanded ? expandedTitle : collapsedTitle,
                systemImage: isExpanded ? "chevron.down" : "chevron.right"
            )
            .font(.subheadline)
        }
        .buttonStyle(.plain)
        .foregroundStyle(.tint)
    }
}

extension Color {
    /// Parses the `#rrggbb` colours the backend stores for tags.
    init?(hex: String) {
        var value = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        if value.hasPrefix("#") { value.removeFirst() }
        guard value.count == 6, let number = UInt32(value, radix: 16) else { return nil }
        self.init(
            red: Double((number >> 16) & 0xFF) / 255,
            green: Double((number >> 8) & 0xFF) / 255,
            blue: Double(number & 0xFF) / 255
        )
    }
}
