package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createincident "github.com/yourorg/boilerplate/internal/command/create_incident"
	resolveincident "github.com/yourorg/boilerplate/internal/command/resolve_incident"
	getincident "github.com/yourorg/boilerplate/internal/query/get_incident_by_id"
	listincidents "github.com/yourorg/boilerplate/internal/query/list_incidents"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// IncidentHandler menangani HTTP request untuk domain Incident.
type IncidentHandler struct {
	BaseHandler
}

var incidentSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"severity":    "severity",
		"status":      "status",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewIncidentHandler membuat instance IncidentHandler baru.
func NewIncidentHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *IncidentHandler {
	return &IncidentHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Incident.
func (h *IncidentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/bcp/incidents", func(r chi.Router) {
		r.Get("/", h.ListIncidents)
		r.Post("/", h.CreateIncident)
		r.Get("/{id}", h.GetIncidentByID)
		r.Put("/{id}/resolve", h.ResolveIncident)
	})
}

// ListIncidents godoc
// @Summary      List incidents
// @Tags         incidents
// @Produce      json
// @Success      200  {object}  listincidents.Result
// @Router       /api/v1/bcp/incidents [get]
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, incidentSortConfig)
	result, err := querybus.Dispatch[*listincidents.Result](r.Context(), h.queryBus, listincidents.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetIncidentByID godoc
// @Summary      Get incident by ID
// @Tags         incidents
// @Produce      json
// @Param        id   path      string  true  "Incident ID"
// @Success      200  {object}  getincident.Result
// @Router       /api/v1/bcp/incidents/{id} [get]
func (h *IncidentHandler) GetIncidentByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getincident.Result](r.Context(), h.queryBus, getincident.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateIncident godoc
// @Summary      Create incident
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/bcp/incidents [post]
func (h *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Severity        string   `json:"severity"`
		IncidentType    string   `json:"incident_type"`
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		AffectedSystems []string `json:"affected_systems"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createincident.Command{
		Severity:        req.Severity,
		IncidentType:    req.IncidentType,
		Title:           req.Title,
		Description:     req.Description,
		AffectedSystems: req.AffectedSystems,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ResolveIncident godoc
// @Summary      Resolve incident
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Incident ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/bcp/incidents/{id}/resolve [put]
func (h *IncidentHandler) ResolveIncident(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		RootCause      string    `json:"root_cause"`
		Resolution     string    `json:"resolution"`
		ResolutionTime int       `json:"resolution_time"`
		ResolvedBy     uuid.UUID `json:"resolved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), resolveincident.Command{
		ID:             id,
		RootCause:      req.RootCause,
		Resolution:     req.Resolution,
		ResolutionTime: req.ResolutionTime,
		ResolvedBy:     req.ResolvedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}
