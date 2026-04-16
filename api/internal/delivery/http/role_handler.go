package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createrolecmd "github.com/yourorg/boilerplate/internal/command/create_role"
	assignrolecmd "github.com/yourorg/boilerplate/internal/command/assign_role"
	"github.com/yourorg/boilerplate/internal/domain/role"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// RoleHandler menangani HTTP request untuk manajemen role.
type RoleHandler struct {
	BaseHandler
	roleReadRepo    role.RoleReadRepository
	userRoleReadRepo role.UserRoleReadRepository
}

// NewRoleHandler membuat instance RoleHandler baru.
func NewRoleHandler(
	roleReadRepo role.RoleReadRepository,
	userRoleReadRepo role.UserRoleReadRepository,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
) *RoleHandler {
	return &RoleHandler{
		BaseHandler:      NewBaseHandler(cb, qb),
		roleReadRepo:     roleReadRepo,
		userRoleReadRepo: userRoleReadRepo,
	}
}

// RegisterRoutes mendaftarkan semua route untuk manajemen role.
func (h *RoleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/roles", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	r.Route("/api/v1/users/{id}/roles", func(r chi.Router) {
		r.Post("/", h.AssignRole)
		r.Get("/", h.GetUserRoles)
		r.Delete("/{roleId}", h.RevokeRole)
	})
}

// List godoc
// @Summary      List roles
// @Tags         roles
// @Produce      json
// @Router       /api/v1/roles [get]
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	s, ok := scope.ScopeFromContext(r.Context())
	if !ok {
		h.RespondError(w, http.StatusForbidden, "scope organisasi tidak teridentifikasi")
		return
	}

	roles, err := h.roleReadRepo.List(r.Context(), s)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data role")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"data": roles,
	})
}

// GetByID godoc
// @Summary      Get role by ID
// @Tags         roles
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Router       /api/v1/roles/{id} [get]
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, ok := scope.ScopeFromContext(r.Context())
	if !ok {
		h.RespondError(w, http.StatusForbidden, "scope organisasi tidak teridentifikasi")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	rl, err := h.roleReadRepo.GetByID(r.Context(), s, id)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "role tidak ditemukan")
		return
	}

	h.RespondJSON(w, http.StatusOK, rl)
}

// Create godoc
// @Summary      Create role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Router       /api/v1/roles [post]
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Code        string   `json:"code"`
		Description string   `json:"description"`
		RoleType    string   `json:"role_type"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createrolecmd.Command{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		RoleType:    req.RoleType,
		Permissions: req.Permissions,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update role
// @Tags         roles
// @Accept       json
// @Param        id   path      string  true  "Role ID"
// @Router       /api/v1/roles/{id} [put]
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		RoleType    string   `json:"role_type"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete role
// @Tags         roles
// @Param        id   path      string  true  "Role ID"
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	_ = id // placeholder until soft-delete command is dispatched
}

// AssignRole godoc
// @Summary      Assign role to user
// @Tags         roles
// @Accept       json
// @Param        id     path      string  true  "User ID"
// @Router       /api/v1/users/{id}/roles [post]
func (h *RoleHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "user id tidak valid")
		return
	}

	var req struct {
		RoleID string `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "role_id tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), assignrolecmd.Command{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "role berhasil ditetapkan"})
}

// GetUserRoles godoc
// @Summary      Get roles for user
// @Tags         roles
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Router       /api/v1/users/{id}/roles [get]
func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "user id tidak valid")
		return
	}

	s, ok := scope.ScopeFromContext(r.Context())
	if !ok {
		h.RespondError(w, http.StatusForbidden, "scope organisasi tidak teridentifikasi")
		return
	}

	roles, err := h.userRoleReadRepo.GetRolesForUser(r.Context(), userID, s.CompanyID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data role user")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"data": roles,
	})
}

// RevokeRole godoc
// @Summary      Revoke role from user
// @Tags         roles
// @Param        id       path      string  true  "User ID"
// @Param        roleId   path      string  true  "Role ID"
// @Router       /api/v1/users/{id}/roles/{roleId} [delete]
func (h *RoleHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "user id tidak valid")
		return
	}

	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "role id tidak valid")
		return
	}

	_ = userID
	_ = roleID
	w.WriteHeader(http.StatusNoContent)
}
