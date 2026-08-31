import SwiftUI

/// Discusses one article with the configured AI provider, keeping the same
/// stored conversations the previous interface used.
struct ArticleChatView: View {
    let article: Article
    let articleContent: String
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var sessions: [ChatSession] = []
    @State private var sessionID: Int?
    @State private var messages: [ChatMessage] = []
    @State private var draft = ""
    @State private var isSending = false
    @State private var errorMessage: String?
    @State private var expandedThinking: Set<Int> = []

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            transcript
            Divider()
            composer
        }
        .frame(width: 620, height: 620)
        .task { await load() }
    }

    private var header: some View {
        HStack(spacing: 10) {
            Label(t("article.chat.aiChat"), systemImage: "bubble.left.and.text.bubble.right")
                .font(.headline)

            Spacer()

            if !sessions.isEmpty {
                Menu {
                    ForEach(sessions) { session in
                        Button(session.title.isEmpty ? article.title : session.title) {
                            Task { await open(session) }
                        }
                    }
                } label: {
                    Label(t("article.chat.switchSession"), systemImage: "clock.arrow.circlepath")
                }
                .menuStyle(.borderlessButton)
                .frame(width: 40)
            }

            Button(t("article.chat.newChat")) {
                sessionID = nil
                messages = []
            }

            Button(t("common.close")) { dismiss() }
                .keyboardShortcut(.cancelAction)
        }
        .padding(16)
    }

    @ViewBuilder
    private var transcript: some View {
        if messages.isEmpty {
            ContentUnavailableView {
                Label(t("article.chat.aiChat"), systemImage: "sparkles")
            } description: {
                Text(t("article.chat.aiChatWelcome"))
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 14) {
                        ForEach(messages) { message in
                            bubble(for: message).id(message.id)
                        }
                        if let errorMessage {
                            Text(errorMessage)
                                .font(.callout)
                                .foregroundStyle(.red)
                        }
                    }
                    .padding(16)
                }
                .onChange(of: messages.count) { _, _ in
                    guard let last = messages.last else { return }
                    withAnimation { proxy.scrollTo(last.id, anchor: .bottom) }
                }
            }
        }
    }

    private func bubble(for message: ChatMessage) -> some View {
        VStack(alignment: message.isAssistant ? .leading : .trailing, spacing: 6) {
            Text(message.content)
                .textSelection(.enabled)
                .padding(.horizontal, 12)
                .padding(.vertical, 9)
                .background(
                    message.isAssistant
                        ? AnyShapeStyle(.quaternary)
                        : AnyShapeStyle(Color.accentColor.opacity(0.18)),
                    in: RoundedRectangle(cornerRadius: 10, style: .continuous)
                )

            if let thinking = message.thinking {
                Button(
                    expandedThinking.contains(message.id)
                        ? t("article.chat.hideThinking")
                        : t("article.chat.showThinking")
                ) {
                    if expandedThinking.contains(message.id) {
                        expandedThinking.remove(message.id)
                    } else {
                        expandedThinking.insert(message.id)
                    }
                }
                .font(.caption)
                .buttonStyle(.plain)
                .foregroundStyle(.tint)

                if expandedThinking.contains(message.id) {
                    Text(thinking)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .textSelection(.enabled)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: message.isAssistant ? .leading : .trailing)
    }

    private var composer: some View {
        HStack(spacing: 10) {
            TextField(t("article.chat.aiChatInputPlaceholder"), text: $draft, axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .lineLimit(1...4)
                .onSubmit { Task { await send() } }

            if isSending {
                ProgressView().controlSize(.small)
            } else {
                Button {
                    Task { await send() }
                } label: {
                    Image(systemName: "paperplane.fill")
                }
                .disabled(draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(16)
    }

    private func load() async {
        sessions = (try? await viewModel.api.fetchChatSessions(articleID: article.id)) ?? []
        if let latest = sessions.first {
            await open(latest)
        }
    }

    private func open(_ session: ChatSession) async {
        sessionID = session.id
        messages = (try? await viewModel.api.fetchChatMessages(sessionID: session.id)) ?? []
    }

    private func send() async {
        let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty, !isSending else { return }

        isSending = true
        errorMessage = nil
        draft = ""

        // Show what was typed straight away, then replace the placeholder
        // identifier when the exchange is stored.
        let localID = (messages.map(\.id).max() ?? 0) + 1
        messages.append(
            ChatMessage(id: localID, sessionID: sessionID ?? 0, role: "user", content: text)
        )

        let turns = messages.map { ChatRequest.Turn(role: $0.role, content: $0.content) }
        let request = ChatRequest(
            messages: turns,
            sessionID: sessionID,
            articleID: article.id,
            articleTitle: article.title,
            articleURL: article.url,
            articleContent: articleContent,
            isFirstMessage: sessionID == nil
        )

        do {
            let response = try await viewModel.api.sendChatMessage(request)
            sessionID = response.sessionID ?? sessionID
            messages.append(
                ChatMessage(
                    id: localID + 1,
                    sessionID: sessionID ?? 0,
                    role: "assistant",
                    content: response.response,
                    html: response.html,
                    thinking: response.thinking
                )
            )
            if !response.historySaved {
                errorMessage = t("article.chat.historySaveFailed")
            }
            sessions = (try? await viewModel.api.fetchChatSessions(articleID: article.id)) ?? sessions
        } catch {
            errorMessage = "\(t("article.chat.aiChatError")) \(error.localizedDescription)"
        }

        isSending = false
    }
}
