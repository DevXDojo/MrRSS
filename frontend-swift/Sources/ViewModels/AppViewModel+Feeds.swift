import Foundation

extension AppViewModel {
    /// Adds or updates a subscription and reloads what changed.
    @discardableResult
    func saveFeed(_ draft: FeedDraft, isEditing: Bool) async -> Bool {
        do {
            if isEditing {
                try await api.updateFeed(draft)
                statusMessage = t("modal.feed.feedUpdatedSuccess")
            } else {
                try await api.addFeed(draft)
                statusMessage = t("modal.feed.feedAddedSuccess")
            }
            refreshFeeds()
            reloadArticles()
            await loadTags()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func refreshFeed(_ feed: Feed) async {
        do {
            try await api.refreshFeed(id: feed.id)
            statusMessage = t("modal.feed.feedRefreshStarted")
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func testIMAPConnection(_ draft: FeedDraft) async throws -> String {
        try await api.testIMAPConnection(draft)
    }

    // MARK: - Tags

    @discardableResult
    func createTag(name: String, color: String) async -> Bool {
        do {
            _ = try await api.createTag(name: name, color: color)
            statusMessage = t("modal.tag.tagCreated")
            await loadTags()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    @discardableResult
    func updateTag(_ tag: Tag) async -> Bool {
        do {
            try await api.updateTag(tag)
            await loadTags()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func deleteTag(_ tag: Tag) async {
        do {
            try await api.deleteTag(id: tag.id)
            await loadTags()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    /// The feeds carrying one tag.
    func feeds(taggedWith tagID: Int) -> [Feed] {
        feeds.filter { feedTags[$0.id]?.contains(tagID) == true }
    }

    // MARK: - Saved filters

    @discardableResult
    func createSavedFilter(name: String, conditions: [FilterCondition]) async -> Bool {
        do {
            _ = try await api.createSavedFilter(name: name, conditions: conditions)
            statusMessage = t("sidebar.savedFilters.filterSaved")
            await loadSavedFilters()
            return true
        } catch {
            errorMessage = "\(t("sidebar.savedFilters.saveFailed")): \(error.localizedDescription)"
            return false
        }
    }

    @discardableResult
    func updateSavedFilter(_ filter: SavedFilter) async -> Bool {
        do {
            try await api.updateSavedFilter(filter)
            statusMessage = t("sidebar.savedFilters.filterUpdated")
            await loadSavedFilters()
            if selection == .savedFilter(filter.id) {
                reloadArticles()
            }
            return true
        } catch {
            errorMessage = "\(t("sidebar.savedFilters.updateFailed")): \(error.localizedDescription)"
            return false
        }
    }

    func deleteSavedFilter(_ filter: SavedFilter) async {
        do {
            try await api.deleteSavedFilter(id: filter.id)
            statusMessage = t("sidebar.savedFilters.filterDeleted")
            if selection == .savedFilter(filter.id) {
                selection = .filter(.all)
            }
            await loadSavedFilters()
        } catch {
            errorMessage = "\(t("sidebar.savedFilters.deleteFailed")): \(error.localizedDescription)"
        }
    }

    // MARK: - OPML

    func exportOPML(to url: URL) async {
        do {
            let data = try await api.exportOPML()
            try data.write(to: url)
            statusMessage = t("modal.opml.exportSuccess")
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func importOPML(from url: URL) async {
        do {
            let data = try Data(contentsOf: url)
            try await api.importOPML(data: data, filename: url.lastPathComponent)
            statusMessage = t("client.opml.importSuccess")
            refreshFeeds()
            reloadArticles()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
