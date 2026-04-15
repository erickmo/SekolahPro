package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createexample "github.com/yourorg/boilerplate/internal/command/create_example"
	deleteexample "github.com/yourorg/boilerplate/internal/command/delete_example"
	updateexample "github.com/yourorg/boilerplate/internal/command/update_example"
	getexample "github.com/yourorg/boilerplate/internal/query/get_example_by_id"
	listexamples "github.com/yourorg/boilerplate/internal/query/list_examples"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// ExampleHandler menangani HTTP request untuk domain Example.
type ExampleHandler struct {
	BaseHandler
}

var exampleSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"name":       "name",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewExampleHandler membuat instance ExampleHandler baru.
func NewExampleHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *ExampleHandler {
	return &ExampleHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Example.
func (h *ExampleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/examples", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List examples
// @Tags         examples
// @Produce      json
// @Success      200  {object}  listexamples.Result
// @Router       /api/v1/examples [get]
func (h *ExampleHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, exampleSortConfig)
	result, err := querybus.Dispatch[*listexamples.Result](r.Context(), h.queryBus, listexamples.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary      Get example by ID
// @Tags         examples
// @Produce      json
// @Param        id   path      string  true  "Example ID"
// @Success      200  {object}  getexample.Result
// @Router       /api/v1/examples/{id} [get]
func (h *ExampleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getexample.Result](r.Context(), h.queryBus, getexample.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create example
// @Tags         examples
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/examples [post]
func (h *ExampleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createexample.Command{
		Name:        req.Name,
		Description: req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update example
// @Tags         examples
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Example ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/examples/{id} [put]
func (h *ExampleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updateexample.Command{
		ID: id, Name: req.Name, Description: req.Description, IsActive: req.IsActive,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete example
// @Tags         examples
// @Param        id   path      string  true  "Example ID"
// @Success      204
// @Router       /api/v1/examples/{id} [delete]
func (h *ExampleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deleteexample.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
