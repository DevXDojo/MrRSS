package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DeliveryError intentionally contains only a stable safe code. Upstream URLs
// and response bodies can contain bot tokens, so never return or log them.
type DeliveryError struct {
	Code  string
	Retry bool
	After time.Duration
}

func (e *DeliveryError) Error() string { return e.Code }

func Send(ctx context.Context, client *http.Client, ch Channel, text string, now time.Time) error {
	if err := ValidateChannel(ch); err != nil {
		return &DeliveryError{Code: err.Error()}
	}
	endpoint := ch.Webhook
	var payload map[string]interface{}
	switch ch.Provider {
	case "telegram":
		endpoint = "https://api.telegram.org/bot" + ch.Token + "/sendMessage"
		payload = map[string]interface{}{"chat_id": ch.ChatID, "text": text, "link_preview_options": map[string]bool{"is_disabled": true}}
		if ch.ThreadID > 0 {
			payload["message_thread_id"] = ch.ThreadID
		}
	case "discord":
		u, _ := url.Parse(endpoint)
		q := u.Query()
		q.Set("wait", "true")
		u.RawQuery = q.Encode()
		endpoint = u.String()
		payload = map[string]interface{}{"content": text, "allowed_mentions": map[string]interface{}{"parse": []string{}}, "flags": 4}
	case "feishu":
		payload = map[string]interface{}{"msg_type": "text", "content": map[string]string{"text": text}}
		if ch.Secret != "" {
			ts := strconv.FormatInt(now.Unix(), 10)
			mac := hmac.New(sha256.New, []byte(ts+"\n"+ch.Secret))
			payload["timestamp"] = ts
			payload["sign"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &DeliveryError{Code: "payload"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return &DeliveryError{Code: "request"}
	}
	req.Header.Set("Content-Type", "application/json")
	// Never follow redirects with credentials in the URL or request body.
	bounded := *client
	bounded.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := bounded.Do(req)
	if err != nil {
		return &DeliveryError{Code: "network", Retry: true}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return &DeliveryError{Code: "network", Retry: true}
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		after := time.Minute
		if seconds, e := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64); e == nil && seconds > 0 {
			after = time.Duration(seconds * float64(time.Second))
		}
		var rate struct {
			RetryAfter float64 `json:"retry_after"`
			Parameters struct {
				RetryAfter int `json:"retry_after"`
			} `json:"parameters"`
		}
		if json.Unmarshal(raw, &rate) == nil {
			if rate.RetryAfter > 0 {
				after = time.Duration(rate.RetryAfter * float64(time.Second))
			}
			if rate.Parameters.RetryAfter > 0 {
				after = time.Duration(rate.Parameters.RetryAfter) * time.Second
			}
		}
		if after < time.Minute {
			after = time.Minute
		}
		if after > 24*time.Hour {
			after = 24 * time.Hour
		}
		return &DeliveryError{Code: "rate_limit", Retry: true, After: after}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := "rejected"
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			code = "unauthorized"
		}
		return &DeliveryError{Code: code, Retry: resp.StatusCode >= 500}
	}
	switch ch.Provider {
	case "telegram":
		var result struct {
			OK bool `json:"ok"`
		}
		if json.Unmarshal(raw, &result) != nil || !result.OK {
			return &DeliveryError{Code: "rejected"}
		}
	case "feishu":
		var result struct {
			Code       *int `json:"code"`
			StatusCode *int `json:"StatusCode"`
		}
		if json.Unmarshal(raw, &result) != nil || (result.Code == nil && result.StatusCode == nil) {
			return &DeliveryError{Code: "response"}
		}
		if result.Code != nil && *result.Code != 0 || result.StatusCode != nil && *result.StatusCode != 0 {
			return &DeliveryError{Code: "feishu_rejected"}
		}
	case "discord":
		// wait=true must confirm a saved message, not merely an accepted request.
		var result struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &result) != nil || result.ID == "" {
			return &DeliveryError{Code: "response"}
		}
	}
	return nil
}

func deliveryFailure(err error) (string, bool, time.Duration) {
	if e, ok := err.(*DeliveryError); ok {
		return e.Code, e.Retry, e.After
	}
	return "internal", false, 0
}
