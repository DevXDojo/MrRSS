import AVFoundation
import AVKit
import SwiftUI

/// Plays an article's audio in place, with the controls the previous interface
/// offered: play, skip, speed and volume.
struct AudioPlayerBar: View {
    let url: URL

    @State private var player: AVPlayer?
    @State private var isPlaying = false
    @State private var progress: Double = 0
    @State private var duration: Double = 0
    @State private var rate: Float = 1
    @State private var volume: Float = 1
    @State private var timeObserver: Any?
    @State private var errorMessage: String?

    private let rates: [Float] = [0.75, 1, 1.25, 1.5, 2]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 10) {
                Label(t("article.audioPlayer.podcastAudio"), systemImage: "waveform")
                    .font(.callout)
                Spacer()
                Menu("\(rate, specifier: "%g")×") {
                    ForEach(rates, id: \.self) { option in
                        Button("\(option, specifier: "%g")×") { setRate(option) }
                    }
                }
                .menuStyle(.borderlessButton)
                .frame(width: 60)
                .help(t("article.audioPlayer.playbackSpeed"))
            }

            HStack(spacing: 12) {
                Button {
                    skip(by: -10)
                } label: {
                    Image(systemName: "gobackward.10")
                }
                .buttonStyle(.borderless)
                .help(t("article.audioPlayer.skipBackward"))

                Button {
                    togglePlayback()
                } label: {
                    Image(systemName: isPlaying ? "pause.fill" : "play.fill")
                        .font(.title3)
                }
                .buttonStyle(.borderless)
                .help(isPlaying ? t("article.audioPlayer.pause") : t("article.audioPlayer.play"))

                Button {
                    skip(by: 10)
                } label: {
                    Image(systemName: "goforward.10")
                }
                .buttonStyle(.borderless)
                .help(t("article.audioPlayer.skipForward"))

                Slider(value: seekBinding, in: 0...max(duration, 1))
                    .controlSize(.small)

                Text(timeLabel)
                    .font(.caption)
                    .monospacedDigit()
                    .foregroundStyle(.secondary)

                Image(systemName: "speaker.wave.2")
                    .foregroundStyle(.secondary)
                Slider(value: volumeBinding, in: 0...1)
                    .controlSize(.small)
                    .frame(width: 70)
                    .help(t("article.audioPlayer.volume"))
            }

            if let errorMessage {
                Text(errorMessage)
                    .font(.caption)
                    .foregroundStyle(.red)
            }
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 12)
        .background(Color.accentColor.opacity(0.05))
        .onAppear(perform: prepare)
        .onDisappear(perform: teardown)
    }

    // MARK: - Playback

    private func prepare() {
        guard player == nil else { return }
        let player = AVPlayer(url: url)
        player.volume = volume
        self.player = player

        timeObserver = player.addPeriodicTimeObserver(
            forInterval: CMTime(seconds: 0.5, preferredTimescale: 600),
            queue: .main
        ) { time in
            progress = time.seconds
            if let total = player.currentItem?.duration.seconds, total.isFinite {
                duration = total
            }
            if player.currentItem?.status == .failed {
                errorMessage = t("article.audioPlayer.audioPlaybackError")
            }
        }
    }

    private func teardown() {
        if let timeObserver {
            player?.removeTimeObserver(timeObserver)
        }
        timeObserver = nil
        player?.pause()
        player = nil
        isPlaying = false
    }

    private func togglePlayback() {
        guard let player else { return }
        if isPlaying {
            player.pause()
        } else {
            player.play()
            player.rate = rate
        }
        isPlaying.toggle()
    }

    private func skip(by seconds: Double) {
        guard let player else { return }
        let target = max(0, player.currentTime().seconds + seconds)
        player.seek(to: CMTime(seconds: target, preferredTimescale: 600))
    }

    private func setRate(_ newRate: Float) {
        rate = newRate
        if isPlaying {
            player?.rate = newRate
        }
    }

    private var seekBinding: Binding<Double> {
        Binding(
            get: { progress },
            set: { newValue in
                progress = newValue
                player?.seek(to: CMTime(seconds: newValue, preferredTimescale: 600))
            }
        )
    }

    private var volumeBinding: Binding<Double> {
        Binding(
            get: { Double(volume) },
            set: { newValue in
                volume = Float(newValue)
                player?.volume = volume
            }
        )
    }

    private var timeLabel: String {
        "\(AudioPlayerBar.format(progress)) / \(AudioPlayerBar.format(duration))"
    }

    /// Formats seconds as `m:ss`, or `h:mm:ss` for anything past an hour.
    static func format(_ seconds: Double) -> String {
        guard seconds.isFinite, seconds >= 0 else { return "0:00" }
        let total = Int(seconds.rounded())
        let hours = total / 3_600
        let minutes = (total % 3_600) / 60
        let remaining = total % 60
        if hours > 0 {
            return String(format: "%d:%02d:%02d", hours, minutes, remaining)
        }
        return String(format: "%d:%02d", minutes, remaining)
    }
}

/// Shows an article's video. A platform that only offers an embed is loaded in a
/// web view; anything playable directly uses the system player.
struct VideoPlayerBar: View {
    let url: URL

    @State private var isExpanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 10) {
                Label(platformTitle, systemImage: "play.rectangle")
                    .font(.callout)
                Spacer()
                Button(isExpanded ? t("common.close") : t("article.audioPlayer.play")) {
                    isExpanded.toggle()
                }
                .controlSize(.small)
                Button(t("article.videoPlayer.openInPlatform", ["platform": platformName])) {
                    NSWorkspace.shared.open(url)
                }
                .controlSize(.small)
            }

            if isExpanded {
                if let embedURL {
                    WebView(source: .url(embedURL))
                        .frame(height: 320)
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                } else {
                    VideoPlayer(player: AVPlayer(url: url))
                        .frame(height: 320)
                        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
            }
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 12)
        .background(Color.accentColor.opacity(0.05))
    }

    private var platformName: String {
        let host = url.host?.lowercased() ?? ""
        if host.contains("youtube") || host.contains("youtu.be") { return "YouTube" }
        if host.contains("vimeo") { return "Vimeo" }
        if host.contains("bilibili") { return "Bilibili" }
        return host
    }

    private var platformTitle: String {
        t("article.videoPlayer.videoPlayer", ["platform": platformName])
    }

    /// The address that embeds the video, when the platform needs one.
    var embedURL: URL? {
        VideoPlayerBar.embedURL(for: url)
    }

    static func embedURL(for url: URL) -> URL? {
        let host = url.host?.lowercased() ?? ""

        if host.contains("youtu.be") {
            let identifier = url.lastPathComponent
            guard !identifier.isEmpty else { return nil }
            return URL(string: "https://www.youtube.com/embed/\(identifier)")
        }

        if host.contains("youtube") {
            if url.path.hasPrefix("/embed/") { return url }
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            guard let identifier = components?.queryItems?.first(where: { $0.name == "v" })?.value else {
                return nil
            }
            return URL(string: "https://www.youtube.com/embed/\(identifier)")
        }

        if host.contains("vimeo") {
            let identifier = url.lastPathComponent
            guard !identifier.isEmpty, Int(identifier) != nil else { return nil }
            return URL(string: "https://player.vimeo.com/video/\(identifier)")
        }

        return nil
    }
}
