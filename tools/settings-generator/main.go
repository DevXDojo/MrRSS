// Settings Code Generator - generates boilerplate code for settings management
// Usage: go run tools/settings-generator/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SettingsSchema defines the structure of settings schema
type SettingsSchema struct {
	Meta     Meta                  `json:"_meta"`
	Settings map[string]SettingDef `json:"settings"`
}

type Meta struct {
	Version     string `json:"version"`
	Description string `json:"description"`
}

type SettingDef struct {
	Type        string      `json:"type"` // int, string, bool
	Default     interface{} `json:"default"`
	Category    string      `json:"category"`
	Encrypted   bool        `json:"encrypted"`
	FrontendKey string      `json:"frontend_key"`
}

func main() {
	// Read schema file
	schemaData, err := os.ReadFile("internal/config/settings_schema.json")
	if err != nil {
		fmt.Printf("Error reading schema: %v\n", err)
		os.Exit(1)
	}

	var schema SettingsSchema
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 Generating code from schema with %d settings...\n\n", len(schema.Settings))

	// Generate all files
	if err := generateDefaultsJSON(&schema); err != nil {
		fmt.Printf("Error generating defaults.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated config/defaults.json")

	if err := generateInternalDefaultsJSON(&schema); err != nil {
		fmt.Printf("Error generating internal defaults.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated internal/config/defaults.json")

	if err := generateConfigGo(&schema); err != nil {
		fmt.Printf("Error generating config.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated internal/config/config.go")

	if err := generateSettingsKeysGo(&schema); err != nil {
		fmt.Printf("Error generating settings_keys.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated internal/config/settings_keys.go")

	if err := generateSettingsHandlersGo(&schema); err != nil {
		fmt.Printf("Error generating settings_handlers.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated internal/handlers/settings/settings_handlers.go")

	if err := generateFrontendTypes(&schema); err != nil {
		fmt.Printf("Error generating frontend types: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated frontend/src/types/settings.generated.ts")

	if err := generateFrontendComposable(&schema); err != nil {
		fmt.Printf("Error generating frontend composable: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated frontend/src/composables/core/useSettings.generated.ts")

	fmt.Println("\n✨ All files generated successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Review generated files")
	fmt.Println("2. Run 'go build' to verify backend code")
	fmt.Println("3. Run 'cd frontend && npm run build' to verify frontend code")
	fmt.Println("4. Update database/db.go to use config.SettingsKeys()")
	fmt.Println("5. Test the application")
}

func generateDefaultsJSON(schema *SettingsSchema) error {
	defaults := make(map[string]interface{})
	// Use backend snake_case keys for defaults.json
	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		defaults[key] = schema.Settings[key].Default
	}

	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return err
	}

	// Ensure file ends with newline (standard practice)
	data = append(data, '\n')

	return os.WriteFile("config/defaults.json", data, 0644)
}

func generateInternalDefaultsJSON(schema *SettingsSchema) error {
	defaults := make(map[string]interface{})
	// Use backend snake_case keys for defaults.json
	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		defaults[key] = schema.Settings[key].Default
	}

	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return err
	}

	// Ensure file ends with newline (standard practice)
	data = append(data, '\n')

	return os.WriteFile("internal/config/defaults.json", data, 0644)
}

func generateConfigGo(schema *SettingsSchema) error {
	// Build struct fields and switch cases
	var structFields []string
	var switchCases []string

	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Find the maximum field name + type length for alignment
	maxFieldLen := 0
	for _, key := range keys {
		def := schema.Settings[key]
		goKey := toGoFieldName(key)
		goType := toGoType(def.Type)
		fieldLen := len(goKey) + len(goType) + 1 // +1 for space
		if fieldLen > maxFieldLen {
			maxFieldLen = fieldLen
		}
	}

	for _, key := range keys {
		def := schema.Settings[key]
		// Convert key to Go field name
		goKey := toGoFieldName(key)
		goType := toGoType(def.Type)

		// Struct field with JSON tag - aligned
		fieldLen := len(goKey) + len(goType) + 1
		padding := maxFieldLen - fieldLen
		structFields = append(structFields, fmt.Sprintf("\t%s %s%s`json:\"%s\"`", goKey, goType, strings.Repeat(" ", padding), key))

		// Switch case for GetString
		caseStmt := fmt.Sprintf("\tcase \"%s\":", key)
		var returnValue string
		switch def.Type {
		case "int":
			returnValue = fmt.Sprintf("strconv.Itoa(defaults.%s)", goKey)
		case "bool":
			returnValue = fmt.Sprintf("strconv.FormatBool(defaults.%s)", goKey)
		case "string":
			returnValue = fmt.Sprintf("defaults.%s", goKey)
		}
		switchCases = append(switchCases, caseStmt, "\t\treturn "+returnValue)
	}

	tmpl := `// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Package config provides centralized default values for settings.
// The defaults are loaded from config/defaults.json which is shared between
// frontend and backend to ensure consistency.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package config

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed defaults.json
var defaultsJSON []byte

// Defaults holds all default settings values
type Defaults struct {
%s
}

var defaults Defaults

func init() {
	if err := json.Unmarshal(defaultsJSON, &defaults); err != nil {
		panic("failed to parse defaults.json: " + err.Error())
	}
}

// Get returns the loaded defaults
func Get() Defaults {
	return defaults
}

// GetString returns a setting default as a string
func GetString(key string) string {
	switch key {
%s
	default:
		return ""
	}
}
`

	content := fmt.Sprintf(tmpl,
		strings.Join(structFields, "\n"),
		strings.Join(switchCases, "\n"))

	return os.WriteFile("internal/config/config.go", []byte(content), 0644)
}

func generateSettingsKeysGo(schema *SettingsSchema) error {
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var keyStrings []string
	for _, key := range keys {
		keyStrings = append(keyStrings, fmt.Sprintf("\"%s\"", key))
	}

	tmpl := `// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Package config provides settings keys for database initialization
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package config

// SettingsKeys returns all valid setting keys
func SettingsKeys() []string {
	return []string{%s}
}
`

	content := fmt.Sprintf(tmpl, strings.Join(keyStrings, ", "))
	return os.WriteFile("internal/config/settings_keys.go", []byte(content), 0644)
}

func generateSettingsHandlersGo(schema *SettingsSchema) error {
	// Generate GET variables
	var getVars []string
	var jsonFields []string

	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := schema.Settings[key]
		varName := toGoVarName(key)
		if def.Encrypted {
			getVars = append(getVars, fmt.Sprintf("\t\t%s := safeGetEncryptedSetting(h, \"%s\")", varName, key))
		} else {
			getVars = append(getVars, fmt.Sprintf("\t\t%s := safeGetSetting(h, \"%s\")", varName, key))
		}
		jsonFields = append(jsonFields, fmt.Sprintf("\t\t\t\"%s\": %s,", key, varName))
	}

	// Generate POST struct fields and save logic
	var structFields []string
	var saveStatements []string

	// Find maximum field name length for alignment
	maxFieldNameLen := 0
	for _, key := range keys {
		goKey := toGoFieldName(key)
		if len(goKey) > maxFieldNameLen {
			maxFieldNameLen = len(goKey)
		}
	}

	for _, key := range keys {
		def := schema.Settings[key]
		goKey := toGoFieldName(key)
		// Align struct field tags
		padding := maxFieldNameLen - len(goKey)
		structFields = append(structFields, fmt.Sprintf("\t\t%s%s string `json:\"%s\"`", goKey, strings.Repeat(" ", padding), key))

		if def.Encrypted {
			saveStatements = append(saveStatements, fmt.Sprintf("\t\tif err := h.DB.SetEncryptedSetting(\"%s\", req.%s); err != nil {\n\t\t\tlog.Printf(\"Failed to save %s: %%v\", err)\n\t\t\thttp.Error(w, \"Failed to save %s\", http.StatusInternalServerError)\n\t\t\treturn\n\t\t}", key, goKey, key, key))
		} else {
			saveStatements = append(saveStatements, fmt.Sprintf("\t\tif req.%s != \"\" {\n\t\t\th.DB.SetSetting(\"%s\", req.%s)\n\t\t}", goKey, key, goKey))
		}
	}

	tmpl := `package settings

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"MrRSS/internal/handlers/core"
)

// safeGetEncryptedSetting safely retrieves an encrypted setting, returning empty string on error.
// This prevents JSON encoding errors when encrypted data is corrupted or cannot be decrypted.
func safeGetEncryptedSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetEncryptedSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to decrypt setting %%s: %%v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// safeGetSetting safely retrieves a setting, returning empty string on error.
func safeGetSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to retrieve setting %%s: %%v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// sanitizeValue removes control characters that could break JSON encoding.
func sanitizeValue(value string) string {
	// Remove control characters that could break JSON
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters except tab, newline, carriage return
		}
		return r
	}, value)
}

// HandleSettings handles GET and POST requests for application settings.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
func HandleSettings(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
%s
		json.NewEncoder(w).Encode(map[string]string{
%s
		})
	case http.MethodPost:
		var req struct {
%s
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
%s
		// Re-fetch all settings after save to return updated values
%s
		json.NewEncoder(w).Encode(map[string]string{
%s
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
`

	content := fmt.Sprintf(tmpl,
		strings.Join(getVars, "\n"),
		strings.Join(jsonFields, "\n"),
		strings.Join(structFields, "\n"),
		strings.Join(saveStatements, "\n\n"),
		strings.Join(getVars, "\n"),
		strings.Join(jsonFields, "\n"))

	return os.WriteFile("internal/handlers/settings/settings_handlers.go", []byte(content), 0644)
}

func generateFrontendTypes(schema *SettingsSchema) error {
	var fields []string
	// Use backend snake_case keys for frontend types
	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := schema.Settings[key]
		tsType := toTSType(def.Type)
		fields = append(fields, fmt.Sprintf("  %s: %s;", key, tsType))
	}

	tmpl := `// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Auto-generated settings types
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go

export interface SettingsData {
%s
  [key: string]: unknown; // Allow additional properties
}
`

	content := fmt.Sprintf(tmpl, strings.Join(fields, "\n"))
	return os.WriteFile("frontend/src/types/settings.generated.ts", []byte(content), 0644)
}

func generateFrontendComposable(schema *SettingsSchema) error {
	var initFields []string
	var fetchFields []string
	var autoSaveFields []string

	// Sort settings by key for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := schema.Settings[key]

		// Init field - using snake_case keys
		initFields = append(initFields, fmt.Sprintf("    %s: settingsDefaults.%s,", key, key))

		// Fetch field conversion
		switch def.Type {
		case "bool":
			fetchFields = append(fetchFields, fmt.Sprintf("    %s: data.%s === 'true',", key, key))
		case "int":
			fetchFields = append(fetchFields, fmt.Sprintf("    %s: parseInt(data.%s) || settingsDefaults.%s,", key, key, key))
		default:
			fetchFields = append(fetchFields, fmt.Sprintf("    %s: data.%s || settingsDefaults.%s,", key, key, key))
		}

		// Auto-save field - convert to string for backend
		// Skip internal settings (should only be modified by backend)
		if def.Category != "internal" {
			switch def.Type {
			case "bool":
				autoSaveFields = append(autoSaveFields, fmt.Sprintf("    %s: (settingsRef.value.%s ?? settingsDefaults.%s).toString(),", key, key, key))
			case "int":
				autoSaveFields = append(autoSaveFields, fmt.Sprintf("    %s: (settingsRef.value.%s ?? settingsDefaults.%s).toString(),", key, key, key))
			default:
				autoSaveFields = append(autoSaveFields, fmt.Sprintf("    %s: settingsRef.value.%s ?? settingsDefaults.%s,", key, key, key))
			}
		}
	}

	tmpl := `// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Auto-generated settings composable helpers
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
import { type Ref } from 'vue';
import type { SettingsData } from '@/types/settings.generated';
import { settingsDefaults } from '@/config/defaults';

/**
 * Generate the initial settings object with defaults
 * This should be used in useSettings() to initialize the settings ref
 */
export function generateInitialSettings(): SettingsData {
  return {
%s
  } as SettingsData;
}

/**
 * Generate the fetchSettings response parser
 * This should be used in useSettings() fetchSettings() to parse backend data
 */
export function parseSettingsData(data: Record<string, string>): SettingsData {
  return {
%s
  } as SettingsData;
}

/**
 * Generate the auto-save payload
 * This should be used in useSettingsAutoSave.ts to build the save payload
 */
export function buildAutoSavePayload(settingsRef: Ref<SettingsData>): Record<string, string> {
  return {
%s
  }
}
`

	content := fmt.Sprintf(tmpl,
		strings.Join(initFields, "\n"),
		strings.Join(fetchFields, "\n"),
		strings.Join(autoSaveFields, "\n"))

	return os.WriteFile("frontend/src/composables/core/useSettings.generated.ts", []byte(content), 0644)
}

// Helper functions
func toGoFieldName(key string) string {
	parts := strings.Split(key, "_")
	for i := 0; i < len(parts); i++ {
		// Capitalize first letter
		if len(parts[i]) > 0 {
			// For freshrss at start, make it FreshRSS
			if i == 0 && parts[i] == "freshrss" {
				parts[i] = "FreshRSS"
			} else if parts[i] == "ai" && i == 0 {
				// ai_ prefix at start should be AI
				parts[i] = "AI"
			} else if parts[i] == "ai" || parts[i] == "api" || parts[i] == "rss" {
				// Keep AI, API, RSS etc uppercase
				parts[i] = strings.ToUpper(parts[i])
			} else {
				// Capitalize first letter
				parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
			}
		}
	}
	return strings.Join(parts, "")
}

func toGoVarName(key string) string {
	// Handle freshrss specially
	if strings.HasPrefix(key, "freshrss") {
		parts := strings.Split(key, "_")
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 {
				parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
			}
		}
		return strings.Join(parts, "")
	}

	// Handle multi-word keys
	parts := strings.Split(key, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func toGoType(typ string) string {
	switch typ {
	case "int":
		return "int"
	case "bool":
		return "bool"
	case "string":
		return "string"
	default:
		return "string"
	}
}

func toTSType(typ string) string {
	switch typ {
	case "int":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	default:
		return "string"
	}
}
