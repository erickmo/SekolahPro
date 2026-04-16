package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createmobileconfig "github.com/yourorg/boilerplate/internal/command/create_mobile_config"
	deletemobileconfig "github.com/yourorg/boilerplate/internal/command/delete_mobile_config"
	updatemobileconfig "github.com/yourorg/boilerplate/internal/command/update_mobile_config"
	getmobileconfig "github.com/yourorg/boilerplate/internal/query/get_mobile_config_by_id"
	listmobileconfigs "github.com/yourorg/boilerplate/internal/query/list_mobile_configs"
	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// MobileConfigHandler menangani HTTP request untuk domain MobileConfig.
type MobileConfigHandler struct {
	BaseHandler
}

var mobileConfigSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"app_variant":  "app_variant",
		"platform":     "platform",
		"created_at":   "created_at",
		"current_version": "current_version",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewMobileConfigHandler membuat instance MobileConfigHandler baru.
func NewMobileConfigHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *MobileConfigHandler {
	return &MobileConfigHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain MobileConfig.
func (h *MobileConfigHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/mobile-configs", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List mobile configs
// @Tags         mobile-configs
// @Produce      json
// @Param        app_variant  query  string  false  "Filter by app_variant (student, parent, staff, full)"
// @Param        platform     query  string  false  "Filter by platform (android, ios, both)"
// @Success      200  {object}  listmobileconfigs.Result
// @Router       /api/v1/mobile-configs [get]
func (h *MobileConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, mobileConfigSortConfig)

	filters := mobile_config.ListFilters{
		AppVariant: mobile_config.AppVariant(r.URL.Query().Get("app_variant")),
		Platform:   mobile_config.Platform(r.URL.Query().Get("platform")),
	}

	result, err := querybus.Dispatch[*listmobileconfigs.Result](r.Context(), h.queryBus, listmobileconfigs.Query{
		Params:  params,
		Filters: filters,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary      Get mobile config by ID
// @Tags         mobile-configs
// @Produce      json
// @Param        id   path      string  true  "MobileConfig ID"
// @Success      200  {object}  getmobileconfig.Result
// @Router       /api/v1/mobile-configs/{id} [get]
func (h *MobileConfigHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getmobileconfig.Result](r.Context(), h.queryBus, getmobileconfig.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create mobile config
// @Tags         mobile-configs
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/mobile-configs [post]
func (h *MobileConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createmobileconfig.Command{
		AppVariant:      req.AppVariant,
		Platform:        req.Platform,
		MinVersion:      req.MinVersion,
		CurrentVersion:  req.CurrentVersion,
		ForceUpdate:     req.ForceUpdate,
		MaintenanceMode: req.MaintenanceMode,
		FeatureFlags:    req.FeatureFlags,
		APIBaseURL:      req.APIBaseURL,
		ThemeConfig:     req.ThemeConfig,
		OfflineConfig:   req.OfflineConfig,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update mobile config
// @Tags         mobile-configs
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "MobileConfig ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/mobile-configs/{id} [put]
func (h *MobileConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
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
		IsActive        bool            `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updatemobileconfig.Command{
		ID:              id,
		AppVariant:      req.AppVariant,
		Platform:        req.Platform,
		MinVersion:      req.MinVersion,
		CurrentVersion:  req.CurrentVersion,
		ForceUpdate:     req.ForceUpdate,
		MaintenanceMode: req.MaintenanceMode,
		FeatureFlags:    req.FeatureFlags,
		APIBaseURL:      req.APIBaseURL,
		ThemeConfig:     req.ThemeConfig,
		OfflineConfig:   req.OfflineConfig,
		IsActive:        req.IsActive,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete mobile config
// @Tags         mobile-configs
// @Param        id   path      string  true  "MobileConfig ID"
// @Success      204
// @Router       /api/v1/mobile-configs/{id} [delete]
func (h *MobileConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletemobileconfig.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
