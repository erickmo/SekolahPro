// Package list_verification_logs menangani query untuk mengambil daftar BiometricVerificationLog.
package list_verification_logs

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric_log"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_verification_logs"

// Query berisi parameter untuk mengambil daftar BiometricVerificationLog.
type Query struct {
	Params pagination.ListParams
	Filter biometric_log.LogFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu BiometricVerificationLog dalam hasil list.
type Item struct {
	ID                 string  `json:"id"`
	EnrollmentID       string  `json:"enrollment_id"`
	NasabahID          string  `json:"nasabah_id"`
	VerificationResult string  `json:"verification_result"`
	FallbackMethod     *string `json:"fallback_method"`
	DeviceInfo         string  `json:"device_info"`
	IPAddress          string  `json:"ip_address"`
	AttemptedAt        string  `json:"attempted_at"`
	CreatedAt          string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_verification_logs.
type Handler struct {
	repo biometric_log.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo biometric_log.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar BiometricVerificationLog dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list verification logs: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		item := &Item{
			ID:                 e.ID.String(),
			EnrollmentID:       e.EnrollmentID.String(),
			NasabahID:          e.NasabahID.String(),
			VerificationResult: string(e.VerificationResult),
			DeviceInfo:         e.DeviceInfo,
			IPAddress:          e.IPAddress,
			AttemptedAt:        e.AttemptedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:          e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.FallbackMethod != nil {
			v := string(*e.FallbackMethod)
			item.FallbackMethod = &v
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}

// BuildFilter membangun LogFilter dari query params.
func BuildFilter(enrollmentID, nasabahID, verificationResult string) biometric_log.LogFilter {
	f := biometric_log.LogFilter{}
	if enrollmentID != "" {
		if id, err := uuid.Parse(enrollmentID); err == nil {
			f.EnrollmentID = &id
		}
	}
	if nasabahID != "" {
		if id, err := uuid.Parse(nasabahID); err == nil {
			f.NasabahID = &id
		}
	}
	if verificationResult != "" {
		vr := biometric_log.VerificationResult(verificationResult)
		f.VerificationResult = &vr
	}
	return f
}
