package app

import (
	"fmt"
	"log"
	"strings"

	"helix-bot/pkg/ports"
)

// stdLogger wraps log for ports.Logger.
type stdLogger struct {
	prefix string
}

// NewStdLogger returns a Logger that uses log.Default().
func NewStdLogger(prefix string) ports.Logger {
	return &stdLogger{prefix: prefix}
}

func (s *stdLogger) Debug(msg string, kv ...any) {
	log.Println("[DEBUG]", s.prefix, msg, fmtKv(kv))
}

func (s *stdLogger) Info(msg string, kv ...any) {
	log.Println("[INFO]", s.prefix, msg, fmtKv(kv))
}

func (s *stdLogger) Warn(msg string, kv ...any) {
	log.Println("[WARN]", s.prefix, msg, fmtKv(kv))
}

func (s *stdLogger) Error(msg string, kv ...any) {
	log.Println("[ERROR]", s.prefix, msg, fmtKv(kv))
}

func fmtKv(kv []any) string {
	if len(kv) == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(kv); i += 2 {
		if i > 0 {
			b.WriteString(" ")
		}
		if i+1 < len(kv) {
			b.WriteString(fmt.Sprint(kv[i]))
			b.WriteString("=")
			b.WriteString(fmt.Sprint(kv[i+1]))
		} else {
			b.WriteString(fmt.Sprint(kv[i]))
		}
	}
	return b.String()
}
