package service

import (
	"log"
	"strings"
	"time"

	"helix-api/internal/domain"
	"helix-api/internal/repo"
)

type AuditService struct {
	repo repo.AuditRepo
}

type AuditListParams struct {
	Page     int
	PageSize int
	Limit    int
	Offset   int
	Keyword  string
	Action   string
	From     string
	To       string
}

type AuditListResult struct {
	List     []domain.AuditLog
	Total    int
	Page     int
	PageSize int
}

func NewAuditService(repo repo.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Append(entry domain.AuditLog) domain.AuditLog {
	// 防御性编程：审计失败不能影响主业务返回
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[warn] audit append panic: %v", r)
		}
	}()
	return s.repo.Append(entry)
}

func (s *AuditService) ListPage(params AuditListParams) AuditListResult {
	list := s.repo.ListAll()
	filtered := make([]domain.AuditLog, 0, len(list))

	var fromTime *time.Time
	var toTime *time.Time
	if params.From != "" {
		if t, err := time.Parse("2006-01-02", params.From); err == nil {
			fromTime = &t
		}
	}
	if params.To != "" {
		if t, err := time.Parse("2006-01-02", params.To); err == nil {
			end := t.Add(24 * time.Hour)
			toTime = &end
		}
	}

	for _, item := range list {
		if params.Action != "" && item.Action != params.Action {
			continue
		}
		if params.Keyword != "" && !strings.Contains(strings.ToLower(item.OperatorName), strings.ToLower(params.Keyword)) {
			continue
		}
		if fromTime != nil && item.CreatedAt.Before(*fromTime) {
			continue
		}
		if toTime != nil && !item.CreatedAt.Before(*toTime) {
			continue
		}
		filtered = append(filtered, item)
	}

	total := len(filtered)

	page := params.Page
	pageSize := params.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if params.Limit > 0 {
		pageSize = params.Limit
		if params.Offset < 0 {
			params.Offset = 0
		}
		page = params.Offset/pageSize + 1
	}

	start := (page - 1) * pageSize
	if params.Limit > 0 {
		start = params.Offset
	}
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return AuditListResult{
		List:     filtered[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
