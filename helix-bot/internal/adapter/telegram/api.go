package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"helix-bot/pkg/errors"
)

const baseURLPrefix = "https://api.telegram.org/bot"

// baseURL returns https://api.telegram.org/bot<token>/ (token never logged).
// Trims token so env/newline issues don't cause 404.
func baseURL(token string) string {
	return baseURLPrefix + strings.TrimSpace(token) + "/"
}

// tgAPIResponse is the common Telegram API response envelope.
type tgAPIResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage  `json:"result,omitempty"`
	Description string          `json:"description,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Parameters  *tgAPIErrParams `json:"parameters,omitempty"`
}

type tgAPIErrParams struct {
	RetryAfter int `json:"retry_after,omitempty"`
}

// doGet performs GET to method with query params. Logs endpoint, duration, error_code only.
func doGet(ctx context.Context, baseURL string, method string, params url.Values) (result []byte, err error) {
	start := time.Now()
	u := baseURL + method
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	dur := time.Since(start)

	var api tgAPIResponse
	_ = json.Unmarshal(body, &api)
	if resp.StatusCode != http.StatusOK || !api.OK {
		code := api.ErrorCode
		if code == 0 && resp.StatusCode != http.StatusOK {
			code = resp.StatusCode
		}
		desc := api.Description
		if desc == "" && resp.StatusCode != http.StatusOK {
			desc = resp.Status
		}
		logAPI(method, dur, code)
		if api.ErrorCode == 429 {
			retry := 0
			if api.Parameters != nil {
				retry = api.Parameters.RetryAfter
			}
			return nil, errors.RateLimitedErr(code, desc, retry)
		}
		return nil, errors.TelegramAPIErr(code, desc)
	}
	// 成功时不打日志，避免空轮询刷屏；由 Source 在有 updates 时打 batch 日志。
	return api.Result, nil
}

// doPost performs POST to method with JSON body. Logs endpoint, duration, error_code only.
func doPost(ctx context.Context, baseURL string, method string, body map[string]any) (result []byte, err error) {
	start := time.Now()
	u := baseURL + method
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	dur := time.Since(start)

	var api tgAPIResponse
	_ = json.Unmarshal(respBody, &api)
	if resp.StatusCode != http.StatusOK || !api.OK {
		code := api.ErrorCode
		if code == 0 && resp.StatusCode != http.StatusOK {
			code = resp.StatusCode
		}
		desc := api.Description
		if desc == "" && resp.StatusCode != http.StatusOK {
			desc = resp.Status
		}
		logAPI(method, dur, code)
		if api.ErrorCode == 429 {
			retry := 0
			if api.Parameters != nil {
				retry = api.Parameters.RetryAfter
			}
			return nil, errors.RateLimitedErr(code, desc, retry)
		}
		return nil, errors.TelegramAPIErr(code, desc)
	}
	return api.Result, nil
}

// logAPI 仅在失败时打日志（endpoint、耗时、error_code），禁止 token/body。 logs only endpoint, duration, and error_code (no token, no body).
func logAPI(endpoint string, dur time.Duration, errorCode int) {
	if errorCode != 0 {
		fmt.Printf("[telegram] %s duration=%v error_code=%d\n", endpoint, dur.Round(time.Millisecond), errorCode)
		return
	}
	fmt.Printf("[telegram] %s duration=%v\n", endpoint, dur.Round(time.Millisecond))
}

// GetUpdates calls getUpdates (offset, timeout, allowed_updates optional). Returns raw result JSON.
func GetUpdates(ctx context.Context, token string, offset int64, timeoutSec int, allowedUpdates []string) (result []byte, err error) {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("timeout", strconv.Itoa(timeoutSec))
	if len(allowedUpdates) > 0 {
		raw, _ := json.Marshal(allowedUpdates)
		params.Set("allowed_updates", string(raw))
	}
	return doGet(ctx, baseURL(token), "getUpdates", params)
}

// sendMessageResult is the result of sendMessage (Message with message_id).
type sendMessageResult struct {
	MessageID int64 `json:"message_id"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

// SendMessage calls sendMessage (chat_id, text, parse_mode optional). Returns message_id.
func SendMessage(ctx context.Context, token string, chatID int64, text string, parseMode string) (messageID int64, err error) {
	body := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if parseMode != "" {
		body["parse_mode"] = parseMode
	}
	raw, err := doPost(ctx, baseURL(token), "sendMessage", body)
	if err != nil {
		return 0, err
	}
	var res sendMessageResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return 0, err
	}
	return res.MessageID, nil
}
