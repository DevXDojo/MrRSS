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

	if err := generateSettingsBaseGo(&schema); err != nil {
		fmt.Printf("Error generating settings_base.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated internal/handlers/settings/settings_base.go")

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

func generateSettingsBaseGo(schema *SettingsSchema) error {
	// Sort keys for consistent output
	var keys []string
	for key := range schema.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Generate setting definitions
	var settingDefs []string
	for _, key := range keys {
		def := schema.Settings[key]
		settingDefs = append(settingDefs, fmt.Sprintf("\t{Key: \"%s\", Encrypted: %v},", key, def.Encrypted))
	}

	tmpl := `// Package settings provides handlers for application settings management.
// This file contains the base types and utilities for the definition-driven settings system.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package settings

import (
	"MrRSS/internal/handlers/core"
)

// SettingDef defines a single setting's metadata
type SettingDef struct {
	Key       string // Database key (snake_case)
	Encrypted bool   // Whether the value should be encrypted in the database
}

// AllSettings returns all setting definitions in alphabetical order by key.
// This is the single source of truth for all settings.
var AllSettings = []SettingDef{
%s
}

// GetAllSettings reads all settings from the database and returns them as a map.
// Encrypted settings are automatically decrypted.
func GetAllSettings(h *core.Handler) map[string]string {
	result := make(map[string]string, len(AllSettings))

	for _, def := range AllSettings {
		var value string
		if def.Encrypted {
			value = safeGetEncryptedSetting(h, def.Key)
		} else {
			value = safeGetSetting(h, def.Key)
		}
		result[def.Key] = value
	}

	return result
}

// SaveSettings saves settings from a map to the database.
// Empty string values are skipped (to allow partial updates).
// Encrypted settings are automatically encrypted.
func SaveSettings(h *core.Handler, settings map[string]string) error {
	// Create a lookup for encrypted keys
	encryptedKeys := make(map[string]bool, len(AllSettings))
	for _, def := range AllSettings {
		if def.Encrypted {
			encryptedKeys[def.Key] = true
		}
	}

	// Save each setting
	for key, value := range settings {
		if encryptedKeys[key] {
			if err := h.DB.SetEncryptedSetting(key, value); err != nil {
				return err
			}
		} else if value != "" {
			h.DB.SetSetting(key, value)
		}
	}

	return nil
}

// IsEncryptedSetting returns true if the given key is an encrypted setting.
func IsEncryptedSetting(key string) bool {
	for _, def := range AllSettings {
		if def.Key == key {
			return def.Encrypted
		}
	}
	return false
}
`

	content := fmt.Sprintf(tmpl, strings.Join(settingDefs, "\n"))
	return os.WriteFile("internal/handlers/settings/settings_base.go", []byte(content), 0644)
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
