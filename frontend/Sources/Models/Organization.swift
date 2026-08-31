import Foundation

/// A user-defined tag attached to feeds.
struct Tag: Identifiable, Codable, Hashable {
    let id: Int
    var name: String
    var color: String
    var position: Int

    init(id: Int, name: String, color: String = "#3b82f6", position: Int = 0) {
        self.id = id
        self.name = name
        self.color = color
        self.position = position
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(Int.self, forKey: .id)
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        color = try container.decodeIfPresent(String.self, forKey: .color) ?? "#3b82f6"
        position = try container.decodeIfPresent(Int.self, forKey: .position) ?? 0
    }
}

/// One clause of a saved filter, matching the shape the previous frontend stored.
struct FilterCondition: Codable, Hashable, Identifiable {
    var id: Int
    var logic: String
    var negate: Bool
    var field: String
    var `operator`: String
    var value: String
    var values: [String]

    init(
        id: Int = Int(Date().timeIntervalSince1970 * 1_000),
        logic: String = "and",
        negate: Bool = false,
        field: String = "article_title",
        operator: String = "contains",
        value: String = "",
        values: [String] = []
    ) {
        self.id = id
        self.logic = logic
        self.negate = negate
        self.field = field
        self.operator = `operator`
        self.value = value
        self.values = values
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = (try? container.decode(Int.self, forKey: .id)) ?? Int(Date().timeIntervalSince1970 * 1_000)
        logic = try container.decodeIfPresent(String.self, forKey: .logic) ?? "and"
        negate = try container.decodeIfPresent(Bool.self, forKey: .negate) ?? false
        field = try container.decodeIfPresent(String.self, forKey: .field) ?? "article_title"
        `operator` = try container.decodeIfPresent(String.self, forKey: .operator) ?? "contains"
        value = try container.decodeIfPresent(String.self, forKey: .value) ?? ""
        values = try container.decodeIfPresent([String].self, forKey: .values) ?? []
    }
}

/// A saved filter. The backend stores the conditions as a JSON string, so the
/// model decodes and re-encodes that payload itself.
struct SavedFilter: Identifiable, Codable, Hashable {
    let id: Int
    var name: String
    var conditions: [FilterCondition]
    var position: Int

    enum CodingKeys: String, CodingKey {
        case id, name, conditions, position
    }

    init(id: Int, name: String, conditions: [FilterCondition] = [], position: Int = 0) {
        self.id = id
        self.name = name
        self.conditions = conditions
        self.position = position
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(Int.self, forKey: .id)
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        position = try container.decodeIfPresent(Int.self, forKey: .position) ?? 0

        if let nested = try? container.decode([FilterCondition].self, forKey: .conditions) {
            conditions = nested
        } else if let encoded = try? container.decode(String.self, forKey: .conditions),
                  let data = encoded.data(using: .utf8),
                  let decoded = try? JSONDecoder().decode([FilterCondition].self, from: data) {
            conditions = decoded
        } else {
            conditions = []
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(id, forKey: .id)
        try container.encode(name, forKey: .name)
        try container.encode(position, forKey: .position)
        let data = try JSONEncoder().encode(conditions)
        try container.encode(String(data: data, encoding: .utf8) ?? "[]", forKey: .conditions)
    }
}

/// One clause of an automation rule.
struct RuleCondition: Codable, Hashable, Identifiable {
    var id: Int
    var logic: String?
    var negate: Bool
    var field: String
    var `operator`: String
    var value: String
    var values: [String]

    static func empty(id: Int = Int(Date().timeIntervalSince1970 * 1_000)) -> RuleCondition {
        RuleCondition(
            id: id,
            logic: "and",
            negate: false,
            field: "article_title",
            operator: "contains",
            value: "",
            values: []
        )
    }
}

struct AutomationRule: Codable, Hashable, Identifiable {
    var id: Int
    var name: String
    var enabled: Bool
    var conditions: [RuleCondition]
    var actions: [String]

    static func empty(id: Int = Int(Date().timeIntervalSince1970 * 1_000)) -> AutomationRule {
        AutomationRule(id: id, name: "", enabled: true, conditions: [], actions: [])
    }
}

struct RuleApplicationResult: Codable, Equatable {
    let success: Bool
    let affected: Int
}
