package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createbackupstrategy "github.com/yourorg/boilerplate/internal/command/create_backup_strategy"
	deletebackupstrategy "github.com/yourorg/boilerplate/internal/command/delete_backup_strategy"
	updatebackupstrategy "github.com/yourorg/boilerplate/internal/command/update_backup_strategy"
	getbackupstrategy "github.com/yourorg/boilerplate/internal/query/get_backup_strategy_by_id"
	listbackupstrategies "github.com/yourorg/boilerplate/internal/query/list_backup_strategies"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// BCPHandler menangani HTTP request untuk domain BCP (Backup Strategy).
type BCPHandler struct {
	BaseHandler
}

var backupStrategySortConfig = SortConfig{
	AllowedFields: map[string]string{
		"strategy_type": "strategy_type",
		"frequency":     "frequency",
		"created_at":    "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewBCPHandler membuat instance BCPHandler baru.
func NewBCPHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *BCPHandler {
	return &BCPHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain BCP.
func (h *BCPHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/bcp/strategies", func(r chi.Router) {
		r.Get("/", h.ListBackupStrategies)
		r.Post("/", h.CreateBackupStrategy)
		r.Get("/{id}", h.GetBackupStrategyByID)
		r.Put("/{id}", h.UpdateBackupStrategy)
		r.Delete("/{id}", h.DeleteBackupStrategy)
	})
}

// ListBackupStrategies godoc
// @Summary      List backup strategies
// @Tags         bcp
// @Produce      json
// @Success      200  {object}  listbackupstrategies.Result
// @Router       /api/v1/bcp/strategies [get]
func (h *BCPHandler) ListBackupStrategies(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, backupStrategySortConfig)
	result, err := querybus.Dispatch[*listbackupstrategies.Result](r.Context(), h.queryBus, listbackupstrategies.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetBackupStrategyByID godoc
// @Summary      Get backup strategy by ID
// @Tags         bcp
// @Produce      json
// @Param        id   path      string  true  "Backup Strategy ID"
// @Success      200  {object}  getbackupstrategy.Result
// @Router       /api/v1/bcp/strategies/{id} [get]
func (h *BCPHandler) GetBackupStrategyByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getbackupstrategy.Result](r.Context(), h.queryBus, getbackupstrategy.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateBackupStrategy godoc
// @Summary      Create backup strategy
// @Tags         bcp
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/bcp/strategies [post]
func (h *BCPHandler) CreateBackupStrategy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StrategyType      string `json:"strategy_type"`
		Frequency         string `json:"frequency"`
		RetentionDays     int    `json:"retention_days"`
		StorageLocation   string `json:"storage_location"`
		EncryptionEnabled bool   `json:"encryption_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createbackupstrategy.Command{
		StrategyType:      req.StrategyType,
		Frequency:         req.Frequency,
		RetentionDays:     req.RetentionDays,
		StorageLocation:   req.StorageLocation,
		EncryptionEnabled: req.EncryptionEnabled,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateBackupStrategy godoc
// @Summary      Update backup strategy
// @Tags         bcp
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Backup Strategy ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/bcp/strategies/{id} [put]
func (h *BCPHandler) UpdateBackupStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		StrategyType      string `json:"strategy_type"`
		Frequency         string `json:"frequency"`
		RetentionDays     int    `json:"retention_days"`
		StorageLocation   string `json:"storage_location"`
		EncryptionEnabled bool   `json:"encryption_enabled"`
		IsActive          bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updatebackupstrategy.Command{
		ID: id, StrategyType: req.StrategyType, Frequency: req.Frequency,
		RetentionDays: req.RetentionDays, StorageLocation: req.StorageLocation,
		EncryptionEnabled: req.EncryptionEnabled, IsActive: req.IsActive,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteBackupStrategy godoc
// @Summary      Delete backup strategy
// @Tags         bcp
// @Param        id   path      string  true  "Backup Strategy ID"
// @Success      204
// @Router       /api/v1/bcp/strategies/{id} [delete]
func (h *BCPHandler) DeleteBackupStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletebackupstrategy.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
