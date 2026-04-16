// Package list_biometrics menangani query untuk mengambil daftar BiometricEnrollment.
package list_biometrics

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_biometrics"

// Query berisi parameter untuk mengambil daftar BiometricEnrollment.
type Query struct {
	Params pagination.ListParams
	Filter biometric.BiometricFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu BiometricEnrollment dalam hasil list.
type Item struct {
	ID             string `json:"id"`
	NasabahID      string `json:"nasabah_id"`
	BiometricType  string `json:"biometric_type"`
	DeviceInfo     string `json:"device_info"`
	IsActive       bool   `json:"is_active"`
	FailedAttempts int    `json:"failed_attempts"`
	CreatedAt      string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_biometrics.
type Handler struct {
	repo biometric.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo biometric.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar BiometricEnrollment dengan pagination, sorting, dan filter scope organisasi.
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
		return nil, fmt.Errorf("list biometrics: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:             e.ID.String(),
			NasabahID:      e.NasabahID.String(),
			BiometricType:  string(e.BiometricType),
			DeviceInfo:     e.DeviceInfo,
			IsActive:       e.IsActive,
			FailedAttempts: e.FailedAttempts,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}

// ParseFilter membaca filter parameter dari HTTP query string.
func ParseFilter(r interface{ URLQuery() string }) biometric.BiometricFilter {
	// Filter di-parse di HTTP handler layer dan diteruskan via Query struct.
	// Method ini disediakan untuk kenyamanan parsing di handler.
	return biometric.BiometricFilter{}
}

// BuildFilter dari query params string values.
func BuildFilter(nasabahID, biometricType, isActive string) biometric.BiometricFilter {
	f := biometric.BiometricFilter{}
	if nasabahID != "" {
		if id, err := uuid.Parse(nasabahID); err == nil {
			f.NasabahID = &id
		}
	}
	if biometricType != "" {
		bt := biometric.BiometricType(biometricType)
		f.BiometricType = &bt
	}
	if isActive == "true" {
		v := true
		f.IsActive = &v
	} else if isActive == "false" {
		v := false
		f.IsActive = &v
	}
	return f
}
