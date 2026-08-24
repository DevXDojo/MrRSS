package translation

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TencentTranslator implements translation using Tencent Cloud Machine Translation API.
// API Documentation: https://cloud.tencent.com/document/api/551/15619
type TencentTranslator struct {
	SecretID  string
	SecretKey string
	Region    string // Default: ap-guangzhou
	client    *http.Client
	db        DBInterface
}

// NewTencentTranslator creates a new Tencent Cloud Translator instance.
// db is optional - if nil, no proxy will be used
func NewTencentTranslator(secretID, secretKey string) *TencentTranslator {
	return &TencentTranslator{
		SecretID:  secretID,
		SecretKey: secretKey,
		Region:    "ap-guangzhou",
		client:    &http.Client{Timeout: 10 * time.Second},
		db:        nil,
	}
}

// NewTencentTranslatorWithRegion creates a new Tencent Translator with custom region.
func NewTencentTranslatorWithRegion(secretID, secretKey, region string) *TencentTranslator {
	return &TencentTranslator{
		SecretID:  secretID,
		SecretKey: secretKey,
		Region:    region,
		client:    &http.Client{Timeout: 10 * time.Second},
		db:        nil,
	}
}

// NewTencentTranslatorWithDB creates a new Tencent Translator with database for proxy support.
func NewTencentTranslatorWithDB(secretID, secretKey string, db DBInterface) *TencentTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &TencentTranslator{
		SecretID:  secretID,
		SecretKey: secretKey,
		Region:    "ap-guangzhou",
		client:    client,
		db:        db,
	}
}

// NewTencentTranslatorWithAll creates a new Tencent Translator with all options.
func NewTencentTranslatorWithAll(secretID, secretKey, region string, db DBInterface) *TencentTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &TencentTranslator{
		SecretID:  secretID,
		SecretKey: secretKey,
		Region:    region,
		client:    client,
		db:        db,
	}
}

// Translate translates text to the target language using Tencent Cloud TMT API.
func (t *TencentTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Map language code to Tencent format
	tencentLang := mapToTencentLang(targetLang)

	// API endpoint
	apiURL := "https://tmt.tencentcloudapi.com/"

	// Current timestamp
	timestamp := time.Now().Unix()

	// Request payload
	payload := map[string]interface{}{
		"SourceText": text,
		"Source":     "auto",
		"Target":     tencentLang,
		"ProjectId":  0,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tencent request: %w", err)
	}

	// Build request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create tencent request: %w", err)
	}

	// Calculate signature
	authorization, err := t.calculateSignature(timestamp, payloadBytes)
	if err != nil {
		return "", fmt.Errorf("failed to calculate signature: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "tmt.tencentcloudapi.com")
	req.Header.Set("X-TC-Action", "TextTranslate")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Version", "2018-03-21")
	req.Header.Set("X-TC-Region", t.Region)
	req.Header.Set("Authorization", authorization)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("tencent api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tencent api returned status: %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
			TargetText string `json:"TargetText"`
			Source     string `json:"Source"`
			Target     string `json:"Target"`
			RequestId  string `json:"RequestId"`
		} `json:"Response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode tencent response: %w", err)
	}

	if result.Response.Error != nil {
		return "", fmt.Errorf("tencent api error: %s - %s", result.Response.Error.Code, result.Response.Error.Message)
	}

	if result.Response.TargetText != "" {
		return result.Response.TargetText, nil
	}

	return "", fmt.Errorf("no translation found in tencent response")
}

// calculateSignature calculates the TC3-HMAC-SHA256 signature for Tencent Cloud API.
func (t *TencentTranslator) calculateSignature(timestamp int64, payload []byte) (string, error) {
	// Service name
	service := "tmt"

	// Create date string
	dateStr := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// Step 1: Build canonical request
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n",
		"application/json; charset=utf-8",
		"tmt.tencentcloudapi.com")
	signedHeaders := "content-type;host"

	// Hash payload
	payloadHash := sha256Hex(payload)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash)

	// Step 2: Build string to sign
	algorithm := "TC3-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", dateStr, service)
	hashedCanonicalRequest := sha256Hex([]byte(canonicalRequest))

	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	// Step 3: Calculate signature
	secretDate := hmacSHA256(dateStr, []byte("TC3"+t.SecretKey))
	secretService := hmacSHA256(service, secretDate)
	secretSigning := hmacSHA256("tc3_request", secretService)
	signature := hex.EncodeToString(hmacSHA256Raw(stringToSign, secretSigning))

	// Step 4: Build authorization header
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		t.SecretID,
		credentialScope,
		signedHeaders,
		signature)

	return authorization, nil
}

// sha256Hex returns the hexadecimal representation of the SHA-256 hash of data.
func sha256Hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSHA256 returns the HMAC-SHA256 hash of data with the given key.
func hmacSHA256(key string, data []byte) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return h.Sum(nil)
}

// hmacSHA256Raw returns the HMAC-SHA256 hash of string data with the given key.
func hmacSHA256Raw(data string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// mapToTencentLang maps standard language codes to Tencent's format.
func mapToTencentLang(lang string) string {
	langMap := map[string]string{
		"en":    "en",
		"zh":    "zh",
		"zh-TW": "zh-TW",
		"es":    "es",
		"fr":    "fr",
		"de":    "de",
		"ja":    "ja",
		"ko":    "ko",
		"pt":    "pt",
		"ru":    "ru",
		"it":    "it",
		"ar":    "ar",
		"tr":    "tr",
		"th":    "th",
		"vi":    "vi",
		"id":    "id",
		"ms":    "ms",
	}
	if tencentLang, ok := langMap[lang]; ok {
		return tencentLang
	}
	return lang
}
