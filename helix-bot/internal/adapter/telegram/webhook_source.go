package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// WebhookSource implements ports.UpdateSource via Telegram webhook HTTP POST.
// 职责：接收 HTTP POST -> 解析 Update -> 去重 -> 推送到 out channel。
// 不在 handler 内执行业务逻辑，尽量快速返回 200。
type WebhookSource struct {
	Addr   string
	Secret string

	logger     ports.Logger
	dedup      *Deduper
	server     *http.Server
	stopped    chan struct{}
	once       sync.Once
	readyCheck func() bool
}

// NewWebhookSource creates an UpdateSource that listens HTTP webhook.
// addr: HTTP_LISTEN_ADDR，例如 ":4000"
// secret: TELEGRAM_WEBHOOK_SECRET，用于路径校验。
// logger: 标准 logger，可为 nil。
func NewWebhookSource(addr, secret string, logger ports.Logger) *WebhookSource {
	return &WebhookSource{
		Addr:       strings.TrimSpace(addr),
		Secret:     strings.TrimSpace(secret),
		logger:     logger,
		dedup:      NewDeduper(defaultDedupWindow),
		stopped:    make(chan struct{}),
		readyCheck: nil,
	}
}

// SetReadyCheck sets a function used by /readyz to determine readiness.
// It is safe to call after NewWebhookSource and before or after Start.
func (s *WebhookSource) SetReadyCheck(fn func() bool) {
	s.readyCheck = fn
}

// Start starts HTTP server and begins accepting webhook POSTs.
// Handler 仅做：方法校验 -> secret 校验 -> JSON 解码 -> 去重 -> 推送到 out（非阻塞）-> 200 OK。
func (s *WebhookSource) Start(ctx context.Context, out chan<- types.BotUpdate) error {
	if s.Secret == "" {
		return errors.New("webhook source: empty secret")
	}
	if s.Addr == "" {
		s.Addr = ":4000"
	}

	mux := http.NewServeMux()
	path := "/tg/" + s.Secret + "/webhook"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// 避免 panic 造成 HTTP Server 崩溃；只记录固定 reason。
				if s.logger != nil {
					s.logger.Error("webhook handler panic", "reason", "panic_recovered")
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 再次校验 path 中的 secret，防御性编程（即使 mux 注册了精确路径）。
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 3 || parts[0] != "tg" || parts[2] != "webhook" || parts[1] != s.Secret {
			http.NotFound(w, r)
			return
		}

		var u tgUpdate
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&u); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 去重：Seen 返回 true 表示之前处理过，则直接 200 OK。
		if s.dedup != nil && s.dedup.Seen(u.UpdateID) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("ok"))
			return
		}

		nu := normalizeUpdate(u)

		// 不在 handler 内做耗时逻辑；向 out 写入采用非阻塞，满了则丢弃并记一条日志。
		select {
		case out <- nu:
			// pushed
		default:
			if s.logger != nil {
				s.logger.Warn("webhook drop", "reason", "channel_full", "update_id", u.UpdateID)
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	// /healthz: process alive -> 200 OK.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	// /readyz: runtime/source considered ready when readyCheck returns true.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ready := false
		if s.readyCheck != nil {
			ready = s.readyCheck()
		}
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	s.server = &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	// 启动 HTTP Server。
	go func() {
		defer close(s.stopped)
		if s.logger != nil {
			s.logger.Info("webhook server starting", "addr", s.Addr)
		}
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if s.logger != nil {
				s.logger.Error("webhook server error", "reason", "listen_failed")
			}
		}
	}()

	// 当 ctx.Done 时，调用 Stop 以关闭 server。
	go func() {
		select {
		case <-ctx.Done():
			_ = s.Stop()
		case <-s.stopped:
		}
	}()

	return nil
}

// Stop gracefully shuts down HTTP server.
func (s *WebhookSource) Stop() error {
	s.once.Do(func() {
		if s.server != nil {
			_ = s.server.Shutdown(context.Background())
		} else {
			close(s.stopped)
		}
	})
	<-s.stopped
	return nil
}

var _ ports.UpdateSource = (*WebhookSource)(nil)
