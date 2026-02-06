package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	bterrors "helix-bot/pkg/errors"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// PollingTimeout is the default long-poll timeout when not set.
const defaultPollTimeoutSec = 60

const (
	backoffMin   = 1 * time.Second
	backoffMax   = 10 * time.Second
	backoffScale = 2
)

// Source implements ports.UpdateSource via Telegram getUpdates long polling.
type Source struct {
	Token      string
	Timeout    time.Duration
	OffsetFile string // optional; persist offset for restart
	offset     int64
	dedup      *Deduper
	logger     ports.Logger
	stop       chan struct{}
	stopped    chan struct{}
	once       sync.Once
}

// NewSource creates an UpdateSource for long polling.
// Token is trimmed to avoid 404 from trailing newline in .env.
// offsetFilePath: optional; if set, load/save offset so restart continues from last.
// logger: 用于标准化日志（updates_count/offset/latency_ms、backoff_ms/reason）；可为 nil 表示不打日志。
func NewSource(token string, timeout time.Duration, offsetFilePath string, logger ports.Logger) *Source {
	if timeout <= 0 {
		timeout = defaultPollTimeoutSec * time.Second
	}
	return &Source{
		Token:      strings.TrimSpace(token),
		Timeout:    timeout,
		OffsetFile: strings.TrimSpace(offsetFilePath),
		dedup:      NewDeduper(defaultDedupWindow),
		logger:     logger,
		stop:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
}

// loadOffset reads offset from file if configured and file exists.
func (s *Source) loadOffset() {
	if s.OffsetFile == "" {
		return
	}
	b, err := os.ReadFile(s.OffsetFile)
	if err != nil {
		return
	}
	v := strings.TrimSpace(string(b))
	if v == "" {
		return
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return
	}
	s.offset = n
}

// saveOffset 将 offset 写入文件（原子写入：先写 .tmp 再 rename）。
func (s *Source) saveOffset() {
	if s.OffsetFile == "" {
		return
	}
	data := []byte(strconv.FormatInt(s.offset, 10) + "\n")
	tmp := s.OffsetFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		if s.logger != nil {
			s.logger.Warn("poll offset write failed", "path", s.OffsetFile, "reason", "write_failed")
		}
		return
	}
	if err := os.Rename(tmp, s.OffsetFile); err != nil {
		if s.logger != nil {
			s.logger.Warn("poll offset rename failed", "path", s.OffsetFile, "reason", "rename_failed")
		}
		return
	}
}

// Start implements ports.UpdateSource. It starts a goroutine that polls getUpdates and sends to out.
func (s *Source) Start(ctx context.Context, out chan<- types.BotUpdate) error {
	s.loadOffset()

	go func() {
		defer close(out)
		defer close(s.stopped)
		timeoutSec := int(s.Timeout.Seconds())
		if timeoutSec < 1 {
			timeoutSec = defaultPollTimeoutSec
		}
		backoff := backoffMin

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			default:
			}

			start := time.Now()
			raw, err := GetUpdates(ctx, s.Token, s.offset, timeoutSec, nil)
			latency := time.Since(start)

			if err != nil {
				var te *bterrors.TelegramError

				// 429: 优先使用 retry_after，其次走指数退避。
				if errors.Is(err, bterrors.ErrRateLimited) {
					if errors.As(err, &te) && te != nil && te.RetryAfter > 0 {
						sleep := time.Duration(te.RetryAfter) * time.Second
						if sleep > backoffMax {
							sleep = backoffMax
						}
						if s.logger != nil {
							s.logger.Warn("polling backoff", "reason", "rate_limited", "backoff_ms", sleep.Milliseconds(), "retry_after_s", te.RetryAfter)
						}
						time.Sleep(sleep)
					} else {
						if s.logger != nil {
							s.logger.Warn("polling backoff", "reason", "rate_limited", "backoff_ms", backoff.Milliseconds())
						}
						time.Sleep(backoff)
						if backoff < backoffMax {
							backoff *= backoffScale
							if backoff > backoffMax {
								backoff = backoffMax
							}
						}
					}
					backoff = backoffMin
					continue
				}

				// 401/403 等永久错误：记录后停止轮询（不打印 description，避免泄露）。
				if errors.Is(err, bterrors.ErrTelegramAPI) && errors.As(err, &te) && te != nil &&
					(te.Code == 401 || te.Code == 403) {
					if s.logger != nil {
						s.logger.Warn("polling stopped", "reason", "unauthorized", "code", te.Code)
					}
					return
				}

				// Ctrl+C / 优雅退出：context 已取消，直接退出，不打 WARN。
				if errors.Is(err, context.Canceled) || ctx.Err() != nil {
					return
				}

				// 其他临时错误：指数退避；只打 reason，禁止 err.Error()（可能含 URL/token）。
				if s.logger != nil {
					s.logger.Warn("polling backoff", "reason", reasonFromErr(err), "backoff_ms", backoff.Milliseconds())
				}
				time.Sleep(backoff)
				if backoff < backoffMax {
					backoff *= backoffScale
					if backoff > backoffMax {
						backoff = backoffMax
					}
				}
				continue
			}

			var list []tgUpdate
			if err := json.Unmarshal(raw, &list); err != nil {
				if s.logger != nil {
					s.logger.Warn("polling backoff", "reason", "decode_error", "backoff_ms", backoff.Milliseconds())
				}
				time.Sleep(backoff)
				if backoff < backoffMax {
					backoff *= backoffScale
					if backoff > backoffMax {
						backoff = backoffMax
					}
				}
				continue
			}

			backoff = backoffMin

			var maxID int64
			for _, u := range list {
				if u.UpdateID > maxID {
					maxID = u.UpdateID
				}
				if s.dedup != nil && s.dedup.Seen(u.UpdateID) {
					if s.logger != nil {
						s.logger.Debug("skip duplicate", "update_id", u.UpdateID)
					}
					continue
				}
				nu := normalizeUpdate(u)
				select {
				case out <- nu:
				case <-ctx.Done():
					return
				case <-s.stop:
					return
				}
			}

			if len(list) > 0 {
				s.offset = maxID + 1
				s.saveOffset()
				if s.logger != nil {
					s.logger.Info("getUpdates", "updates_count", len(list), "offset", s.offset, "latency_ms", latency.Milliseconds())
				}
			}
		}
	}()
	return nil
}

// Stop implements ports.UpdateSource. Safe to call multiple times.
func (s *Source) Stop() error {
	s.once.Do(func() { close(s.stop) })
	<-s.stopped
	return nil
}

type tgUpdate struct {
	UpdateID      int64       `json:"update_id"`
	Message       *tgMessage  `json:"message,omitempty"`
	CallbackQuery *tgCallback `json:"callback_query,omitempty"`
}

type tgMessage struct {
	MessageID int64   `json:"message_id"`
	Chat      tgChat  `json:"chat"`
	From      *tgUser `json:"from,omitempty"`
	Text      string  `json:"text,omitempty"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type tgUser struct {
	ID int64 `json:"id"`
}

type tgCallback struct {
	ID      string     `json:"id"`
	From    tgUser     `json:"from"`
	Message *tgMessage `json:"message,omitempty"`
	Data    string     `json:"data,omitempty"`
}

func normalizeUpdate(u tgUpdate) types.BotUpdate {
	nu := types.BotUpdate{Raw: u, UpdateID: u.UpdateID}
	if u.Message != nil {
		nu.Message = &types.BotMessage{
			ChatID:    u.Message.Chat.ID,
			MessageID: u.Message.MessageID,
			Text:      u.Message.Text,
		}
		if u.Message.From != nil {
			nu.Message.UserID = u.Message.From.ID
		}
	}
	if u.CallbackQuery != nil {
		nu.CallbackQuery = &types.BotCallbackQuery{
			ID:     u.CallbackQuery.ID,
			Data:   u.CallbackQuery.Data,
			UserID: u.CallbackQuery.From.ID,
		}
		if u.CallbackQuery.Message != nil {
			nu.CallbackQuery.ChatID = u.CallbackQuery.Message.Chat.ID
			nu.CallbackQuery.MessageID = u.CallbackQuery.Message.MessageID
		}
	}
	return nu
}

var _ ports.UpdateSource = (*Source)(nil)
