// Package get_mobile_config_by_id menangani query untuk mengambil MobileAppConfig berdasarkan ID.
package get_mobile_config_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_mobile_config_by_id"

// Query berisi parameter untuk mengambil MobileAppConfig berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID              uuid.UUID                  `json:"id"`
	AppVariant      mobile_config.AppVariant   `json:"app_variant"`
	Platform        mobile_config.Platform     `json:"platform"`
	MinVersion      string                     `json:"min_version"`
	CurrentVersion  string                     `json:"current_version"`
	ForceUpdate     bool                       `json:"force_update"`
	MaintenanceMode bool                       `json:"maintenance_mode"`
	FeatureFlags    mobile_config.FeatureFlags  `json:"feature_flags"`
	APIBaseURL      string                     `json:"api_base_url"`
	ThemeConfig     mobile_config.ThemeConfig   `json:"theme_config"`
	OfflineConfig   mobile_config.OfflineConfig `json:"offline_config"`
	IsActive        bool                       `json:"is_active"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
}

// Handler menangani Query get_mobile_config_by_id.
type Handler struct {
	repo mobile_config.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo mobile_config.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil MobileAppConfig berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get mobile_config by id: %w", err)
	}

	return &Result{
		ID:              entity.ID,
		AppVariant:      entity.AppVariant,
		Platform:        entity.Platform,
		MinVersion:      entity.MinVersion,
		CurrentVersion:  entity.CurrentVersion,
		ForceUpdate:     entity.ForceUpdate,
		MaintenanceMode: entity.MaintenanceMode,
		FeatureFlags:    entity.FeatureFlags,
		APIBaseURL:      entity.APIBaseURL,
		ThemeConfig:     entity.ThemeConfig,
		OfflineConfig:   entity.OfflineConfig,
		IsActive:        entity.IsActive,
		CreatedAt:       entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
