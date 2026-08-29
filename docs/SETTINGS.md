# Settings Management System

## Table of Contents

- [Quick Start](#quick-start)
- [Overview](#overview)
- [How to Add a New Setting](#how-to-add-a-new-setting)
- [Complete Example](#complete-example)
- [Reference](#reference)

---

## Quick Start

**Adding a new setting is now simple - just 3 steps:**

### 1. Add to Schema

Edit `internal/config/settings_schema.json`:

```json
"your_setting_key": {
  "type": "bool",              // or "int", "string"
  "default": false,
  "category": "general",
  "encrypted": false,
  "frontend_key": "your_setting_key"  // snake_case (same as key)
}
```

### 2. Generate Code

```bash
go run tools/settings-generator/main.go
```

### 3. Generate the client catalogue

```bash
python3 tools/settings-swift/generate.py
```

The settings window builds itself from the catalogue, so the new setting appears
without any view being edited. Add its wording to the translation catalogue so
it reads properly.

That's it! See [Complete Example](#complete-example) for a detailed walkthrough.

---

## Overview

The settings system has been optimized to use **schema-driven code generation**. Instead of manually editing 11+ files, you now only need to edit **1 file** and run the code generator.

### Before vs After

**Old Way (Deprecated):**

- Edit 11 files manually
- ~50-100 lines of repetitive code
- High chance of copy-paste errors
- 30-45 minutes of work

**New Way (Current):**

- Edit 1 file (5 lines)
- Run 1 command
- Add UI and translations (optional)
- 10-15 minutes of work

**Result:** ~90% reduction in development time, near-zero error risk

### What Gets Generated

After running the generator, these files are **automatically created/updated**:

- ✅ `internal/config/config.go` - Go struct and `GetString()` function
- ✅ `internal/config/settings_keys.go` - Settings keys array for DB init
- ✅ `internal/handlers/settings/settings_handlers.go` - GET/POST API handlers
- ✅ `config/defaults.json` - Shipped defaults (snake_case)
- ✅ `internal/config/defaults.json` - Backend defaults (snake_case)

**Important:** All generated files are sorted alphabetically to minimize diff changes when adding new settings.

### Naming Convention

**Settings are addressed by their snake_case schema key everywhere**:

- ✅ `viewModel.setting("ai_api_key")` (correct)
- ❌ `viewModel.setting("aiAPIKey")` (incorrect)

This holds across the Go configuration, the API payloads and the macOS client.

---

## How to Add a New Setting

### Step 1: Define in Schema

Edit `internal/config/settings_schema.json`, add to the `settings` object:

```json
"your_new_setting": {
  "type": "bool",              // Type: "bool", "int", or "string"
  "default": false,            // Default value
  "category": "general",       // Category (see reference below)
  "encrypted": false,          // Set to true for sensitive data
  "frontend_key": "your_new_setting"  // Use same snake_case as key
}
```

**Schema Properties:**

| Property | Type | Description |
| -------- | ---- | ----------- |
| `type` | string | Required: `"bool"`, `"int"`, or `"string"` |
| `default` | mixed | Required: Default value (must match type) |
| `category` | string | Required: See [Categories](#categories) below |
| `encrypted` | boolean | Required: `true` for sensitive data (API keys, passwords) |
| `frontend_key` | string | Required: Use the same snake_case as the key (for reference) |

**Note:** The `frontend_key` is currently for reference only and should match the key in snake_case. The actual frontend implementation uses snake_case property names.

### Step 2: Run the Code Generator

```bash
go run tools/settings-generator/main.go
```

**Output:**

```plaintext
🔧 Generating code from schema with 66 settings...

✓ Generated config/defaults.json
✓ Generated internal/config/defaults.json
✓ Generated internal/config/config.go
✓ Generated internal/config/settings_keys.go
✓ Generated internal/handlers/settings/settings_handlers.go

✨ All files generated successfully!
```

Then generate the client's settings catalogue:

```bash
python3 tools/settings-swift/generate.py
```

**Output:**

```plaintext
wrote frontend-swift/Sources/Models/SettingsCatalog.generated.swift with 99 settings
```

The catalogue carries the key, the pane, the control, the default and the
translation keys, so the settings window picks the new setting up on its own.

### Step 3: Add Translations (Recommended)

The generator looks for wording in the client's catalogue,
`frontend-swift/Sources/Localization/LocalizationTables.swift`, using the same
keys the previous frontend used. Add the label and, optionally, a description:

```json
"setting.general.yourNewSetting": "Your New Setting",
"setting.general.yourNewSettingDesc": "What this setting does"
```

Add the Chinese wording to the `chineseSimplified` table in the same file. If no
match is found the generator falls back to a readable form of the key, which is
visible but not translated, so it is worth adding.

For wording only the client needs, use `ClientStrings.swift` and a `client.`
prefix instead.

### Step 4: Adjust the Generator (Only If Needed)

Most settings need nothing further. Reach for
`tools/settings-swift/generate.py` when:

- **The translation key cannot be derived from the name**: add it to
  `LABEL_OVERRIDES`
- **The setting belongs on a different pane than its schema category**: add it
  to `PANE_OVERRIDES`
- **The value comes from a fixed list**: add the options to `CHOICES`, each with
  the stored value and its translation key
- **The setting is internal**: add it to `HIDDEN` so it stays out of the window

Then regenerate.

### Step 5: Implement Feature Logic (Optional)

If the setting changes how the client behaves, read it where the behaviour
lives. Settings are strings on the wire, so use the typed accessors:

```swift
// Boolean
if viewModel.boolSetting("your_new_setting", default: true) {
    // ...
}

// String, with the schema default as the fallback
let mode = viewModel.setting("default_view_mode", default: "rendered")

// Number
let size = Int(viewModel.setting("content_font_size", default: "16")) ?? 16
```

Nothing needs to listen for a change: `AppViewModel.settings` is published, so a
view reading it redraws when the value is saved. A setting that changes
something outside SwiftUI — the language, for instance — is applied in
`loadSettings()`.

### Step 6: Test

```bash
# Backend
go build ./...
go test ./internal/config/...

# Client
swift build --package-path frontend-swift
swift test --package-path frontend-swift --filter SettingsCatalogTests

# Or run both and check the settings window
./frontend-swift/run.sh
```

---

## Complete Example

Let's walk through adding a complete setting from start to finish.

### Goal: Add "Auto-Collapse Sidebar" Setting

We want to add a setting that automatically collapses the sidebar on startup.

### Step 1: Define in Schema

Edit `internal/config/settings_schema.json`:

```json
"auto_collapse_sidebar": {
  "type": "bool",
  "default": false,
  "category": "general",
  "encrypted": false,
  "frontend_key": "auto_collapse_sidebar"
}
```

**Why these values?**

- `type: "bool"` - It's a toggle/checkbox setting
- `default: false` - Most users want sidebar expanded by default
- `category: "general"` - It's a general UI preference
- `encrypted: false` - Not sensitive data
- `frontend_key: "auto_collapse_sidebar"` - Use snake_case (same as key)

### Step 2: Generate Code

```bash
go run tools/settings-generator/main.go
```

**What was generated?**

1. **`internal/config/config.go`**
   - Added `AutoCollapseSidebar bool` field to Defaults struct
   - Added switch case for `GetString("auto_collapse_sidebar")`

2. **`internal/config/settings_keys.go`**
   - Added `"auto_collapse_sidebar"` to keys array

3. **`internal/handlers/settings/settings_handlers.go`**
   - Added GET: `autoCollapseSidebar, _ := h.DB.GetSetting("auto_collapse_sidebar")`
   - Added JSON response: `"auto_collapse_sidebar": autoCollapseSidebar`
   - Added POST field: `AutoCollapseSidebar string \`json:"auto_collapse_sidebar"\``
   - Added save logic: `if req.AutoCollapseSidebar != "" { h.DB.SetSetting(...) }`

4. **`frontend-swift/Sources/Models/SettingsCatalog.generated.swift`** (after
   running the Swift generator)
   - Added a `SettingDefinition` with the key, pane, control and default

6. **`config/defaults.json` & `internal/config/defaults.json`**
   - Added: `"auto_collapse_sidebar": false`

### Step 3: Add Translations

In `frontend-swift/Sources/Localization/LocalizationTables.swift`, add to the
English table:

```json
"setting.general.autoCollapseSidebar": "Auto Collapse Sidebar",
"setting.general.autoCollapseSidebarDesc": "Automatically collapse the sidebar when the app starts"
```

And to the Chinese table:

```json
"setting.general.autoCollapseSidebar": "自动折叠侧边栏",
"setting.general.autoCollapseSidebarDesc": "应用启动时自动折叠侧边栏"
```

### Step 4: Regenerate the Catalogue

```bash
python3 tools/settings-swift/generate.py
```

The setting now appears on the General pane as a switch, with its description
underneath. Nothing else needs editing.

### Step 5: Implement Feature Logic

Where the sidebar is built:

```swift
if viewModel.boolSetting("auto_collapse_sidebar") {
    columnVisibility = .detailOnly
}
```

### Step 6: Test

#### Manual Testing

1. Open Settings → General
2. Find "Auto Collapse Sidebar" setting
3. Toggle it on
4. Close and reopen the app
5. ✅ Verify sidebar is collapsed on startup
6. Toggle it off
7. Close and reopen the app
8. ✅ Verify sidebar is expanded on startup

#### Check Database

```sql
SELECT * FROM settings WHERE key = 'auto_collapse_sidebar';
```

Should show:

```plaintext
key                     | value
------------------------+-------
auto_collapse_sidebar   | true
```

#### Verify API

**GET** `/api/settings`:

```bash
curl http://localhost:5343/api/settings
```

Should include:

```json
{
  "auto_collapse_sidebar": "true"
}
```

**POST** `/api/settings`:

```bash
curl -X POST http://localhost:5343/api/settings \
  -H "Content-Type: application/json" \
  -d '{"auto_collapse_sidebar": "false"}'
```

Should return `200 OK`.

### Complete Checklist

- [x] Schema added to `settings_schema.json`
- [x] Code generator ran successfully
- [x] Backend compiles without errors
- [x] Frontend compiles without errors
- [x] English translations added
- [x] Chinese translations added
- [x] UI component added (Toggle in GeneralSettings)
- [x] Feature logic implemented (sidebar collapse)
- [x] Setting appears in settings modal
- [x] Setting saves to database
- [x] Setting loads on startup
- [x] API GET returns correct value
- [x] API POST saves value correctly

---

## Reference

### Type Mapping

| Schema Type | Go Type | Client Control | Example |
| ----------- | ------- | -------------- | ------- |
| `"bool"` | `bool` | Switch | `true`, `false` |
| `"int"` | `int` | Number field | `30`, `500` |
| `"string"` | `string` | Text field, secure field, or picker | `"en"`, `"openai"` |

A string setting becomes a picker when the generator's `CHOICES` table lists its
options, and a secure field when the schema marks it encrypted or the key ends
in `_key`, `_password` or `_secret`.

### Categories

Use the appropriate category for your setting:

| Category | Description | Example Settings |
| -------- | ----------- | ---------------- |
| `general` | General app settings | theme, language, shortcuts |
| `reading` | Reading/viewing preferences | view mode, hover mark as read, show hidden |
| `translation` | Translation settings | provider, target language, API keys |
| `ai` | AI-related settings | API key, model, prompts, usage limit |
| `summary` | Article summary settings | summary length, trigger mode |
| `storage` | Cache and storage settings | cache size, cleanup, max age |
| `network` | Network and proxy settings | proxy, bandwidth, concurrent refreshes |
| `integrations` | Third-party integrations | Obsidian, FreshRSS |
| `internal` | Internal app state (no UI) | window position, last update |

### Encrypted Settings

For sensitive data (API keys, passwords), set `"encrypted": true`:

```json
"my_api_key": {
  "type": "string",
  "default": "",
  "category": "integrations",
  "encrypted": true,  // ← Important!
  "frontend_key": "my_api_key"
}
```

Encrypted settings are automatically:

- Stored encrypted in the database
- Fetched using `GetEncryptedSetting()` instead of `GetSetting()`
- Saved using `SetEncryptedSetting()` instead of `SetSetting()`

### Key Convention

**Important:** settings are addressed by their snake_case schema key everywhere.

| Schema Key | Read in the client as |
| ---------- | --------------------- |
| `update_interval` | `viewModel.setting("update_interval")` |
| `startup_on_boot` | `viewModel.boolSetting("startup_on_boot")` |
| `deepl_api_key` | `viewModel.setting("deepl_api_key")` |
| `ai_chat_enabled` | `viewModel.boolSetting("ai_chat_enabled")` |

The `frontend_key` in the schema is a hint for the generators when the
translation key cannot be derived from the setting name.

### Quick Examples

**Boolean Setting:**

```json
"enable_feature": {
  "type": "bool",
  "default": true,
  "category": "general",
  "encrypted": false,
  "frontend_key": "enable_feature"
}
```

Read in the client:

```swift
if viewModel.boolSetting("enable_feature", default: true) {
    // ...
}
```

**Integer Setting:**

```json
"max_items": {
  "type": "int",
  "default": 100,
  "category": "storage",
  "encrypted": false,
  "frontend_key": "max_items"
}
```

**String Setting:**

```json
"api_endpoint": {
  "type": "string",
  "default": "https://api.example.com",
  "category": "integrations",
  "encrypted": false,
  "frontend_key": "api_endpoint"
}
```

**Encrypted Setting:**

```json
"api_secret": {
  "type": "string",
  "default": "",
  "category": "integrations",
  "encrypted": true,    // ← Encrypts in DB
  "frontend_key": "api_secret"
}
```

### Event Listeners

For non-internal settings, change events are automatically dispatched. Listen to them like this:

```typescript
window.addEventListener('your-setting-key-changed', (event) => {
  const { value } = event.detail
  console.log('Setting changed to:', value)
})
```

Event name format: `{key in kebab-case}-changed`

Examples:

- `auto_collapse_sidebar` → `auto-collapse-sidebar-changed`
- `ai_api_key` → `ai-api-key-changed`
- `ai_chat_enabled` → `ai-chat-enabled-changed`

### Common Mistakes

❌ **Wrong:**

```json
"type": "boolean",     // Should be "bool"
"category": "General", // Should be lowercase
"frontend_key": "myAPIKey" // Should be snake_case (my_api_key)
```

✅ **Correct:**

```json
"type": "bool",
"category": "general",
"frontend_key": "my_api_key"
```

### Troubleshooting

#### Build Errors

**Problem:** `go build` fails after adding a setting

**Solution:**

1. Check that your `settings_schema.json` has valid JSON syntax (no missing commas)
2. Verify `type` is one of: `"bool"`, `"int"`, `"string"`
3. Verify `category` is a valid category
4. Run generator again: `go run tools/settings-generator/main.go`

#### Frontend Errors

**Problem:** the setting is missing from the client

**Solution:**

1. Make sure you ran both generators
2. Check that `SettingsCatalog.generated.swift` contains the key
3. Rebuild with `swift build --package-path frontend-swift`
4. If the label reads as a raw key, add its wording to the translation catalogue

#### Setting Not Appearing in UI

**Problem:** Toggle doesn't show in settings modal

**Solution:**

1. Check that you added the `<SettingItem>` component
2. Verify the translation keys match
3. Check browser console for errors
4. Try hard refresh (Ctrl+Shift+R)

#### Setting Not Saving

**Problem:** Toggle changes but resets on restart

**Solution:**

1. Open browser DevTools → Network tab
2. Check if POST to `/api/settings` is sent
3. Check response status (should be 200 OK)
4. Check database directly via SQLite browser
5. Verify the key name matches in schema

---

## Migration from Old System

If you have existing manually-written settings code:

1. ✅ Ensure all settings are defined in `internal/config/settings_schema.json`
2. ✅ Run the generator: `go run tools/settings-generator/main.go`
3. ✅ Review and commit the generated files
4. ✅ Delete any manual setting-related code that's now replaced

The generated code is compatible with the existing database and API.

---

## Best Practices

1. **Use descriptive names** - `enable_auto_sync` not `eas`
2. **Choose appropriate types** - Use `bool` for toggles, `int` for numbers
3. **Set sensible defaults** - What should the setting be for new users?
4. **Add translations** - Always add both English and Chinese
5. **Use categories** - This helps organize the settings UI
6. **Encrypt sensitive data** - API keys, passwords, tokens
7. **Test after adding** - Run the app and verify the setting works
8. **Document complex settings** - Add comments if behavior is non-obvious
9. **Use snake_case** - Frontend uses snake_case consistently (not camelCase)
10. **Keep frontend_key same as key** - The `frontend_key` should match the setting key

---

## Summary

**Old workflow:** Edit 11 files, ~100 lines of code, high chance of errors

**New workflow:** Edit 1 file (5 lines), run 1 command, done!

This optimization:

- ✅ Reduces development time by ~90%
- ✅ Eliminates copy-paste errors
- ✅ Ensures consistency between frontend and backend
- ✅ Maintains type safety automatically
- ✅ Makes adding new settings trivial
- ✅ Uses snake_case throughout (simpler than camelCase)

Happy coding! 🚀
