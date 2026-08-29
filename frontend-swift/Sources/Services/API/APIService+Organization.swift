import Foundation

extension APIService {
    // MARK: - Tags

    func fetchTags() async throws -> [Tag] {
        try await get("tags")
    }

    func createTag(name: String, color: String) async throws -> Tag {
        try await postDecoding("tags", jsonBody: ["name": name, "color": color])
    }

    func updateTag(_ tag: Tag) async throws {
        try await post(
            "tags/update",
            jsonBody: [
                "id": tag.id,
                "name": tag.name,
                "color": tag.color,
                "position": tag.position
            ]
        )
    }

    func deleteTag(id: Int) async throws {
        try await post("tags/delete", jsonBody: ["id": id])
    }

    func reorderTag(id: Int, newPosition: Int) async throws {
        try await post("tags/reorder", jsonBody: ["id": id, "new_position": newPosition])
    }

    // MARK: - Saved filters

    func fetchSavedFilters() async throws -> [SavedFilter] {
        try await get("saved-filters")
    }

    func createSavedFilter(name: String, conditions: [FilterCondition]) async throws -> SavedFilter {
        let encoded = try JSONEncoder().encode(conditions)
        return try await postDecoding(
            "saved-filters",
            jsonBody: [
                "name": name,
                "conditions": String(data: encoded, encoding: .utf8) ?? "[]"
            ]
        )
    }

    func updateSavedFilter(_ filter: SavedFilter) async throws {
        let encoded = try JSONEncoder().encode(filter.conditions)
        try await send(
            "saved-filters/filter",
            method: "PUT",
            queryItems: [URLQueryItem(name: "id", value: String(filter.id))],
            jsonBody: [
                "id": filter.id,
                "name": filter.name,
                "conditions": String(data: encoded, encoding: .utf8) ?? "[]"
            ]
        )
    }

    func deleteSavedFilter(id: Int) async throws {
        try await send(
            "saved-filters/filter",
            method: "DELETE",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func reorderSavedFilter(id: Int, newPosition: Int) async throws {
        try await post("saved-filters/reorder", jsonBody: ["id": id, "new_position": newPosition])
    }

    // MARK: - Rules

    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult {
        try await postJSON("rules/apply", body: rule)
    }

    // MARK: - Settings

    func fetchSettings() async throws -> [String: String] {
        try await get("settings")
    }

    func updateSettings(_ settings: [String: String]) async throws {
        try await sendJSONReturningData("settings", body: settings)
    }
}
