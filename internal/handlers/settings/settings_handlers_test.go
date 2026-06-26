package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
)

func setupHandlerWithDB(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return core.NewHandler(db, nil, nil, nil)
}

func TestHandleSettings_GET(t *testing.T) {
	h := setupHandlerWithDB(t)

	// Set a custom value
	h.DB.SetSetting("language", "xx-YY")

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if data["language"] != "xx-YY" {
		t.Fatalf("expected language xx-YY, got %s", data["language"])
	}
}

func TestHandleSettings_GETJSONContractForFlutterClient(t *testing.T) {
	h := setupHandlerWithDB(t)

	values := map[string]string{
		"language":                         "zh-CN",
		"theme":                            "dark",
		"update_interval":                  "45",
		"refresh_mode":                     "intelligent",
		"default_view_mode":                "webpage",
		"translation_enabled":              "true",
		"translation_only_mode":            "false",
		"target_language":                  "en",
		"translation_provider":             "deepl",
		"deepl_api_key":                    "secret-key",
		"deepl_endpoint":                   "https://api-free.deepl.com/v2/translate",
		"baidu_app_id":                     "baidu-app",
		"baidu_secret_key":                 "baidu-secret",
		"microsoft_api_key":                "ms-key",
		"microsoft_endpoint":               "https://api.cognitive.microsofttranslator.com",
		"microsoft_region":                 "eastasia",
		"tencent_secret_id":                "tencent-id",
		"tencent_secret_key":               "tencent-key",
		"tencent_region":                   "ap-shanghai",
		"ai_translation_profile_id":        "profile-1",
		"ai_translation_prompt":            "Translate precisely.",
		"custom_translation_enabled":       "true",
		"custom_translation_name":          "Acme Translate",
		"custom_translation_endpoint":      "https://translate.example.com",
		"custom_translation_method":        "POST",
		"custom_translation_headers":       `{"X-API-Key":"secret"}`,
		"custom_translation_body_template": `{"text":"{{text}}"}`,
		"custom_translation_response_path": "data.text",
		"custom_translation_lang_mapping":  `{"zh":"zh-CN","en":"en-US"}`,
		"custom_translation_timeout":       "15",
	}
	for key, value := range values {
		if IsEncryptedSetting(key) {
			if err := h.DB.SetEncryptedSetting(key, value); err != nil {
				t.Fatalf("set encrypted setting %s: %v", key, err)
			}
			continue
		}
		h.DB.SetSetting(key, value)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assertSettingString(t, data, "language", "zh-CN")
	assertSettingString(t, data, "theme", "dark")
	assertSettingString(t, data, "update_interval", "45")
	assertSettingString(t, data, "refresh_mode", "intelligent")
	assertSettingString(t, data, "default_view_mode", "webpage")
	assertSettingString(t, data, "translation_enabled", "true")
	assertSettingString(t, data, "translation_only_mode", "false")
	assertSettingString(t, data, "target_language", "en")
	assertSettingString(t, data, "translation_provider", "deepl")
	assertSettingString(t, data, "deepl_api_key", "secret-key")
	assertSettingString(t, data, "deepl_endpoint", "https://api-free.deepl.com/v2/translate")
	assertSettingString(t, data, "baidu_app_id", "baidu-app")
	assertSettingString(t, data, "baidu_secret_key", "baidu-secret")
	assertSettingString(t, data, "microsoft_api_key", "ms-key")
	assertSettingString(t, data, "microsoft_endpoint", "https://api.cognitive.microsofttranslator.com")
	assertSettingString(t, data, "microsoft_region", "eastasia")
	assertSettingString(t, data, "tencent_secret_id", "tencent-id")
	assertSettingString(t, data, "tencent_secret_key", "tencent-key")
	assertSettingString(t, data, "tencent_region", "ap-shanghai")
	assertSettingString(t, data, "ai_translation_profile_id", "profile-1")
	assertSettingString(t, data, "ai_translation_prompt", "Translate precisely.")
	assertSettingString(t, data, "custom_translation_enabled", "true")
	assertSettingString(t, data, "custom_translation_name", "Acme Translate")
	assertSettingString(t, data, "custom_translation_endpoint", "https://translate.example.com")
	assertSettingString(t, data, "custom_translation_method", "POST")
	assertSettingString(t, data, "custom_translation_headers", `{"X-API-Key":"secret"}`)
	assertSettingString(t, data, "custom_translation_body_template", `{"text":"{{text}}"}`)
	assertSettingString(t, data, "custom_translation_response_path", "data.text")
	assertSettingString(t, data, "custom_translation_lang_mapping", `{"zh":"zh-CN","en":"en-US"}`)
	assertSettingString(t, data, "custom_translation_timeout", "15")

	for _, key := range []string{"language", "theme", "update_interval", "translation_enabled"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("settings response missing required Flutter bootstrap key %q", key)
		}
	}
}

func TestHandleSettings_POST(t *testing.T) {
	h := setupHandlerWithDB(t)

	payload := map[string]string{
		"update_interval":                  "15",
		"translation_enabled":              "true",
		"translation_only_mode":            "true",
		"target_language":                  "ja",
		"translation_provider":             "google",
		"deepl_api_key":                    "deadbeef",
		"deepl_endpoint":                   "https://api-free.deepl.com/v2/translate",
		"baidu_app_id":                     "baidu-app",
		"baidu_secret_key":                 "baidu-secret",
		"microsoft_api_key":                "ms-key",
		"microsoft_endpoint":               "https://api.cognitive.microsofttranslator.com",
		"microsoft_region":                 "eastasia",
		"tencent_secret_id":                "tencent-id",
		"tencent_secret_key":               "tencent-key",
		"tencent_region":                   "ap-shanghai",
		"ai_translation_profile_id":        "profile-1",
		"ai_translation_prompt":            "Translate precisely.",
		"custom_translation_enabled":       "true",
		"custom_translation_name":          "Acme Translate",
		"custom_translation_endpoint":      "https://translate.example.com",
		"custom_translation_method":        "GET",
		"custom_translation_headers":       `{"X-API-Key":"secret"}`,
		"custom_translation_body_template": `{"text":"{{text}}"}`,
		"custom_translation_response_path": "data.text",
		"custom_translation_lang_mapping":  `{"zh":"zh-CN","en":"en-US"}`,
		"custom_translation_timeout":       "15",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify settings saved
	v, _ := h.DB.GetSetting("update_interval")
	if v != "15" {
		t.Fatalf("expected update_interval 15, got %s", v)
	}

	v2, _ := h.DB.GetSetting("translation_enabled")
	if v2 != "true" {
		t.Fatalf("expected translation_enabled true, got %s", v2)
	}
	v3, _ := h.DB.GetSetting("translation_only_mode")
	if v3 != "true" {
		t.Fatalf("expected translation_only_mode true, got %s", v3)
	}
	v4, _ := h.DB.GetSetting("target_language")
	if v4 != "ja" {
		t.Fatalf("expected target_language ja, got %s", v4)
	}
	v5, _ := h.DB.GetSetting("translation_provider")
	if v5 != "google" {
		t.Fatalf("expected translation_provider google, got %s", v5)
	}
	v6, _ := h.DB.GetSetting("deepl_endpoint")
	if v6 != "https://api-free.deepl.com/v2/translate" {
		t.Fatalf("expected deepl_endpoint saved, got %s", v6)
	}
	v7, _ := h.DB.GetSetting("baidu_app_id")
	if v7 != "baidu-app" {
		t.Fatalf("expected baidu_app_id saved, got %s", v7)
	}
	decBaidu, err := h.DB.GetEncryptedSetting("baidu_secret_key")
	if err != nil {
		t.Fatalf("GetEncryptedSetting baidu_secret_key error: %v", err)
	}
	if decBaidu != "baidu-secret" {
		t.Fatalf("expected baidu_secret_key decrypted to be baidu-secret, got %s", decBaidu)
	}
	decMicrosoft, err := h.DB.GetEncryptedSetting("microsoft_api_key")
	if err != nil {
		t.Fatalf("GetEncryptedSetting microsoft_api_key error: %v", err)
	}
	if decMicrosoft != "ms-key" {
		t.Fatalf("expected microsoft_api_key decrypted to be ms-key, got %s", decMicrosoft)
	}
	v8, _ := h.DB.GetSetting("microsoft_endpoint")
	if v8 != "https://api.cognitive.microsofttranslator.com" {
		t.Fatalf("expected microsoft_endpoint saved, got %s", v8)
	}
	v9, _ := h.DB.GetSetting("microsoft_region")
	if v9 != "eastasia" {
		t.Fatalf("expected microsoft_region saved, got %s", v9)
	}
	v10, _ := h.DB.GetSetting("tencent_secret_id")
	if v10 != "tencent-id" {
		t.Fatalf("expected tencent_secret_id saved, got %s", v10)
	}
	decTencent, err := h.DB.GetEncryptedSetting("tencent_secret_key")
	if err != nil {
		t.Fatalf("GetEncryptedSetting tencent_secret_key error: %v", err)
	}
	if decTencent != "tencent-key" {
		t.Fatalf("expected tencent_secret_key decrypted to be tencent-key, got %s", decTencent)
	}
	v11, _ := h.DB.GetSetting("tencent_region")
	if v11 != "ap-shanghai" {
		t.Fatalf("expected tencent_region saved, got %s", v11)
	}
	v12, _ := h.DB.GetSetting("ai_translation_profile_id")
	if v12 != "profile-1" {
		t.Fatalf("expected ai_translation_profile_id saved, got %s", v12)
	}
	v13, _ := h.DB.GetSetting("ai_translation_prompt")
	if v13 != "Translate precisely." {
		t.Fatalf("expected ai_translation_prompt saved, got %s", v13)
	}
	for key, want := range map[string]string{
		"custom_translation_enabled":       "true",
		"custom_translation_name":          "Acme Translate",
		"custom_translation_endpoint":      "https://translate.example.com",
		"custom_translation_method":        "GET",
		"custom_translation_headers":       `{"X-API-Key":"secret"}`,
		"custom_translation_body_template": `{"text":"{{text}}"}`,
		"custom_translation_response_path": "data.text",
		"custom_translation_lang_mapping":  `{"zh":"zh-CN","en":"en-US"}`,
		"custom_translation_timeout":       "15",
	} {
		got, _ := h.DB.GetSetting(key)
		if got != want {
			t.Fatalf("expected %s %s, got %s", key, want, got)
		}
	}

	// Encrypted key should be retrievable via GetEncryptedSetting
	dec, err := h.DB.GetEncryptedSetting("deepl_api_key")
	if err != nil {
		t.Fatalf("GetEncryptedSetting error: %v", err)
	}
	if dec != "deadbeef" {
		t.Fatalf("expected deepl_api_key decrypted to be deadbeef, got %s", dec)
	}
}

func assertSettingString(t *testing.T, data map[string]string, key, want string) {
	t.Helper()

	got, ok := data[key]
	if !ok {
		t.Fatalf("settings response missing %q", key)
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
