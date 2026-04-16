// Package delete_mobile_config menangani command untuk soft-delete MobileAppConfig.
package delete_mobile_config

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_mobile_config"

// Command berisi data untuk menghapus MobileAppConfig.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_mobile_config.
type Handler struct {
	repo mobile_config.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo mobile_config.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle melakukan soft delete dan memvalidasi scope.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.ID == uuid.Nil {
		return fmt.Errorf("ID tidak boleh kosong")
	}
	return h.repo.Delete(ctx, s, cmd.ID)
}
