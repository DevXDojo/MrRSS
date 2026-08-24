package translation

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

// BaiduTranslator implements translation using the Baidu Translate API.
type BaiduTranslator struct {
	AppID     string
	SecretKey string
	client    *http.Client
	db        DBInterface
}

// NewBaiduTranslator creates a new Baidu translator with the given credentials.
// db is optional - if nil, no proxy will be used
func NewBaiduTranslator(appID, secretKey string) *BaiduTranslator {
	return &BaiduTranslator{
		AppID:     appID,
		SecretKey: secretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
		db:        nil,
	}
}

// NewBaiduTranslatorWithDB creates a new Baidu translator with database for proxy support
func NewBaiduTranslatorWithDB(appID, secretKey string, db DBInterface) *BaiduTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		// Fallback to default client if proxy creation fails
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &BaiduTranslator{
		AppID:     appID,
		SecretKey: secretKey,
		client:    client,
		db:        db,
	}
}

// Translate translates text to the target language using Baidu Translate API.
func (t *BaiduTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Baidu API uses different language codes
	baiduLang := mapToBaiduLang(targetLang)

	// Generate cryptographically secure random salt
	n, err := rand.Int(rand.Reader, big.NewInt(1000000000))
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	salt := n.String()

	// Generate sign: md5(appid+q+salt+key)
	// Note: MD5 is used here because it's required by the Baidu Translate API specification.
	// This is not for security purposes but for API signature verification.
	signStr := t.AppID + text + salt + t.SecretKey
	hash := md5.Sum([]byte(signStr))
	sign := hex.EncodeToString(hash[:])

	// Build request URL
	apiURL := "https://fanyi-api.baidu.com/api/trans/vip/translate"
	data := url.Values{}
	data.Set("q", text)
	data.Set("from", "auto")
	data.Set("to", baiduLang)
	data.Set("appid", t.AppID)
	data.Set("salt", salt)
	data.Set("sign", sign)

	resp, err := t.client.PostForm(apiURL, data)
	if err != nil {
		return "", fmt.Errorf("baidu api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baidu api returned status: %d", resp.StatusCode)
	}

	var result struct {
		ErrorCode   string `json:"error_code"`
		ErrorMsg    string `json:"error_msg"`
		TransResult []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode baidu response: %w", err)
	}

	if result.ErrorCode != "" && result.ErrorCode != "52000" {
		return "", fmt.Errorf("baidu api error: %s - %s", result.ErrorCode, result.ErrorMsg)
	}

	if len(result.TransResult) > 0 {
		return result.TransResult[0].Dst, nil
	}

	return "", fmt.Errorf("no translation found in baidu response")
}

// mapToBaiduLang maps standard language codes to Baidu's language codes.
func mapToBaiduLang(lang string) string {
	langMap := map[string]string{
		"en":    "en",
		"zh":    "zh",
		"zh-TW": "cht",
		"es":    "spa",
		"fr":    "fra",
		"de":    "de",
		"ja":    "jp",
		"ko":    "kor",
		"pt":    "pt",
		"ru":    "ru",
		"it":    "it",
		"ar":    "ara",
	}
	if baiduLang, ok := langMap[lang]; ok {
		return baiduLang
	}
	return lang
}
