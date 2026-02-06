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
	Token         string
	Timeout       time.Duration
	OffsetFile    string // optional; persist offset for restart
	offset        int64
	dedup         *dedupSet
	stop          chan struct{}
	stopped       chan struct{}
	once          sync.Once
}

// NewSource creates an UpdateSource for long polling.
// Token is trimmed to avoid 404 from trailing newline in .env.
// offsetFilePath: optional; if set, load/save offset so restart continues from last (avoids re-scanning old updates).
func NewSource(token string, timeout time.Duration, offsetFilePath string) *Source {
	if timeout <= 0 {
		timeout = defaultPollTimeoutSec * time.Second
	}
	return &Source{
		Token:      strings.TrimSpace(token),
		Timeout:    timeout,
		OffsetFile: strings.TrimSpace(offsetFilePath),
		dedup:      newDedupSet(defaultDedupWindow),
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

// saveOffset writes offset to file if configured (atomic: write to .tmp then rename).
func (s *Source) saveOffset() {
	if s.OffsetFile == "" {
		return
	}
	data := []byte(strconv.FormatInt(s.offset, 10) + "\n")
	tmp := s.OffsetFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return
	}
	_ = os.Rename(tmp, s.OffsetFile)
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

			raw, err := GetUpdates(ctx, s.Token, s.offset, timeoutSec, nil)
			if err != nil {
				if errors.Is(err, bterrors.ErrRateLimited) {
					backoff = backoffMin
					time.Sleep(5 * time.Second)
				} else {
					time.Sleep(backoff)
					if backoff < backoffMax {
						backoff *= backoffScale
						if backoff > backoffMax {
							backoff = backoffMax
						}
					}
				}
				continue
			}

			var list []tgUpdate
			if err := json.Unmarshal(raw, &list); err != nil {
				time.Sleep(backoff)
				if backoff < backoffMax {
					backoff *= backoffScale
					if backoff > backoffMax {
						backoff = backoffMax
					}
				}
				continue
			}

			backoff = backoffMin // reset on any success (getUpdates + unmarshal)

			var maxID int64
			for _, u := range list {
				if u.UpdateID > maxID {
					maxID = u.UpdateID
				}
				if s.dedup.Seen(u.UpdateID) {
					continue
				}
				s.dedup.Add(u.UpdateID)
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
	UpdateID      int64        `json:"update_id"`
	Message       *tgMessage   `json:"message,omitempty"`
	CallbackQuery *tgCallback  `json:"callback_query,omitempty"`
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
	ID      string    `json:"id"`
	From    tgUser    `json:"from"`
	Message *tgMessage `json:"message,omitempty"`
	Data    string    `json:"data,omitempty"`
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
