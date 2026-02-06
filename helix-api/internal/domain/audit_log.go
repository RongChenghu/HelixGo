package domain

import "time"

type AuditLog struct {
	ID           int64
	Action       string
	Method       string
	Path         string
	Status       int
	OperatorID   string
	OperatorName string
	IP           string
	UserAgent    string
	TraceID      string
	CreatedAt    time.Time
}
