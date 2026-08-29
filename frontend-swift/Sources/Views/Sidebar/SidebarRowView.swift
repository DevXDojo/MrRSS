import AppKit

/// A source list row: icon, title, and an unread count on the trailing edge.
final class SidebarRowView: NSTableCellView {
    private let icon = NSImageView()
    private let title = NSTextField(labelWithString: "")
    private let badge = NSTextField(labelWithString: "")
    private var iconTask: Task<Void, Never>?

    static let reuseIdentifier = NSUserInterfaceItemIdentifier("SidebarRow")

    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)

        icon.translatesAutoresizingMaskIntoConstraints = false
        icon.imageScaling = .scaleProportionallyUpOrDown
        icon.setContentHuggingPriority(.required, for: .horizontal)

        title.translatesAutoresizingMaskIntoConstraints = false
        title.lineBreakMode = .byTruncatingTail
        title.font = .systemFont(ofSize: NSFont.systemFontSize)
        title.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)

        badge.translatesAutoresizingMaskIntoConstraints = false
        badge.font = .systemFont(ofSize: NSFont.smallSystemFontSize)
        badge.textColor = .secondaryLabelColor
        badge.alignment = .right
        badge.setContentHuggingPriority(.required, for: .horizontal)
        badge.setContentCompressionResistancePriority(.required, for: .horizontal)

        addSubview(icon)
        addSubview(title)
        addSubview(badge)
        imageView = icon
        textField = title

        NSLayoutConstraint.activate([
            icon.leadingAnchor.constraint(equalTo: leadingAnchor),
            icon.centerYAnchor.constraint(equalTo: centerYAnchor),
            icon.widthAnchor.constraint(equalToConstant: 16),
            icon.heightAnchor.constraint(equalToConstant: 16),

            title.leadingAnchor.constraint(equalTo: icon.trailingAnchor, constant: 6),
            title.centerYAnchor.constraint(equalTo: centerYAnchor),

            badge.leadingAnchor.constraint(greaterThanOrEqualTo: title.trailingAnchor, constant: 6),
            badge.trailingAnchor.constraint(equalTo: trailingAnchor),
            badge.centerYAnchor.constraint(equalTo: centerYAnchor)
        ])
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override func prepareForReuse() {
        super.prepareForReuse()
        iconTask?.cancel()
        iconTask = nil
    }

    func configure(with node: SidebarNode) {
        title.stringValue = node.title
        badge.stringValue = node.badge > 0 ? String(node.badge) : ""
        applyIcon(for: node)
    }

    private func applyIcon(for node: SidebarNode) {
        iconTask?.cancel()
        iconTask = nil

        switch node.kind {
        case .group:
            icon.image = nil
        case .filter(let filter):
            icon.contentTintColor = .controlAccentColor
            icon.image = NSImage(systemSymbolName: filter.icon, accessibilityDescription: nil)
        case .folder:
            icon.contentTintColor = .controlAccentColor
            icon.image = NSImage(systemSymbolName: "folder", accessibilityDescription: nil)
        case .feed(let feed):
            applyFeedIcon(for: feed)
        }
    }

    private func applyFeedIcon(for feed: Feed) {
        icon.contentTintColor = NSColor(FeedAccentColor.color(for: feed))
        icon.image = NSImage(systemSymbolName: "dot.radiowaves.left.and.right", accessibilityDescription: nil)

        guard let address = feed.iconURL, let url = URL(string: address) else { return }

        if let cached = RemoteImageLoader.shared.cachedImage(for: url, maxPixelSize: 32) {
            show(cached)
            return
        }

        iconTask = Task { [weak self] in
            let image = await RemoteImageLoader.shared.image(for: url, maxPixelSize: 32)
            guard !Task.isCancelled, let image, let self else { return }
            self.show(image)
        }
    }

    private func show(_ image: NSImage) {
        icon.contentTintColor = nil
        icon.image = image
    }
}
