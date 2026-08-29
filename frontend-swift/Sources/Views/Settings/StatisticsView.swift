import SwiftUI

/// Reading statistics for a chosen period, with the same figures the previous
/// interface reported.
struct StatisticsView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var period = "week"
    @State private var offset = 0
    @State private var summary: StatisticsSummary?
    @State private var allTime: [String: Int] = [:]
    @State private var isLoading = false

    private let metrics: [(key: String, titleKey: String)] = [
        ("articles_read", "setting.statistic.articlesRead"),
        ("articles_viewed", "setting.statistic.articlesViewed"),
        ("articles_favorited", "setting.statistic.articlesFavorited"),
        ("ai_summaries", "setting.statistic.aiSummaries"),
        ("ai_chats", "setting.statistic.aiChats")
    ]

    var body: some View {
        Form {
            Section {
                Picker(t("setting.statistic.statistics"), selection: $period) {
                    Text(t("setting.statistic.byWeek")).tag("week")
                    Text(t("setting.statistic.byMonth")).tag("month")
                    Text(t("setting.statistic.byYear")).tag("year")
                    Text(t("setting.statistic.allTime")).tag("all")
                }
                .pickerStyle(.segmented)

                if let summary {
                    HStack {
                        Button {
                            offset -= 1
                        } label: {
                            Image(systemName: "chevron.left")
                        }
                        .disabled(!summary.hasPrevious)

                        Spacer()
                        Text(summary.displayLabel)
                            .font(.headline)
                        Spacer()

                        Button {
                            offset += 1
                        } label: {
                            Image(systemName: "chevron.right")
                        }
                        .disabled(!summary.hasNext)
                    }
                }
            } header: {
                Text(t("setting.statistic.description"))
            }

            Section(t("setting.statistic.statistics")) {
                if isLoading {
                    ProgressView()
                } else {
                    ForEach(metrics, id: \.key) { metric in
                        LabeledContent(t(metric.titleKey)) {
                            Text("\(summary?.totals[metric.key] ?? 0)")
                                .monospacedDigit()
                        }
                    }
                }
            }

            Section(t("setting.statistic.allTime")) {
                ForEach(metrics, id: \.key) { metric in
                    LabeledContent(t(metric.titleKey)) {
                        Text("\(allTime[metric.key] ?? 0)")
                            .monospacedDigit()
                    }
                }
            }
        }
        .formStyle(.grouped)
        .task(id: "\(period)-\(offset)") { await load() }
        .onChange(of: period) { _, _ in offset = 0 }
    }

    private func load() async {
        isLoading = true
        defer { isLoading = false }
        summary = try? await viewModel.api.fetchStatistics(period: period, offset: offset)
        allTime = (try? await viewModel.api.fetchAllTimeStatistics()) ?? [:]
    }
}
