// Package create_mobile_config menangani command untuk membuat MobileAppConfig baru.
package create_mobile_config

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_mobile_config"

// SemverRegex memvalidasi format versi semver (x.y.z).
var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// Command berisi data yang dibutuhkan untuk membuat MobileAppConfig baru.
type Command struct {
	AppVariant      string          `json:"app_variant"`
	Platform        string          `json:"platform"`
	MinVersion      string          `json:"min_version"`
	CurrentVersion  string          `json:"current_version"`
	ForceUpdate     bool            `json:"force_update"`
	MaintenanceMode bool            `json:"maintenance_mode"`
	FeatureFlags    json.RawMessage `json:"feature_flags"`
	APIBaseURL      string          `json:"api_base_url"`
	ThemeConfig     json.RawMessage `json:"theme_config"`
	OfflineConfig   json.RawMessage `json:"offline_config"`
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_mobile_config.
type Handler struct {
	repo     mobile_config.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo mobile_config.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan ke repository,
// dan mempublikasikan domain event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	appVariant := mobile_config.AppVariant(cmd.AppVariant)
	if !mobile_config.ValidAppVariants[appVariant] {
		return mobile_config.ErrInvalidAppVariant
	}

	platform := mobile_config.Platform(cmd.Platform)
	if !mobile_config.ValidPlatforms[platform] {
		return mobile_config.ErrInvalidPlatform
	}

	if cmd.MinVersion != "" && !semverRegex.MatchString(cmd.MinVersion) {
		return mobile_config.ErrInvalidVersion
	}
	if cmd.CurrentVersion == "" || !semverRegex.MatchString(cmd.CurrentVersion) {
		return mobile_config.ErrInvalidVersion
	}

	if cmd.APIBaseURL == "" {
		return mobile_config.ErrEmptyAPIBaseURL
	}
	if len(cmd.APIBaseURL) > 255 {
		return mobile_config.ErrAPIBaseURLTooLong
	}

	var featureFlags mobile_config.FeatureFlags
	if cmd.FeatureFlags != nil {
		if err := featureFlags.Scan(cmd.FeatureFlags); err != nil {
			return fmt.Errorf("invalid feature_flags: %w", err)
		}
	} else {
		featureFlags = make(mobile_config.FeatureFlags)
	}

	var themeConfig mobile_config.ThemeConfig
	if cmd.ThemeConfig != nil {
		if err := themeConfig.Scan(cmd.ThemeConfig); err != nil {
			return fmt.Errorf("invalid theme_config: %w", err)
		}
	} else {
		themeConfig = make(mobile_config.ThemeConfig)
	}

	var offlineConfig mobile_config.OfflineConfig
	if cmd.OfflineConfig != nil {
		if err := offlineConfig.Scan(cmd.OfflineConfig); err != nil {
			return fmt.Errorf("invalid offline_config: %w", err)
		}
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &mobile_config.MobileAppConfig{
		ID:              id,
		AppVariant:      appVariant,
		Platform:        platform,
		MinVersion:      cmd.MinVersion,
		CurrentVersion:  cmd.CurrentVersion,
		ForceUpdate:     cmd.ForceUpdate,
		MaintenanceMode: cmd.MaintenanceMode,
		FeatureFlags:    featureFlags,
		APIBaseURL:      cmd.APIBaseURL,
		ThemeConfig:     themeConfig,
		OfflineConfig:   offlineConfig,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save mobile_config: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, mobile_config.MobileConfigCreatedEvent{
		TenantID:   s.TenantID,
		CompanyID:  s.CompanyID,
		ID:         id,
		AppVariant: appVariant,
		Platform:   platform,
	}); err != nil {
		return fmt.Errorf("publish mobile_config.created: %w", err)
	}

	return nil
}
