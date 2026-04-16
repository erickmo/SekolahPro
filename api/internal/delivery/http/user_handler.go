package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	registercmd "github.com/yourorg/boilerplate/internal/command/register_user"
	"github.com/yourorg/boilerplate/internal/domain/user"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/middleware"
	"github.com/yourorg/boilerplate/pkg/querybus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// UserHandler menangani HTTP request untuk manajemen user.
type UserHandler struct {
	BaseHandler
	userReadRepo  user.ReadRepository
	userWriteRepo user.WriteRepository
}

// NewUserHandler membuat instance UserHandler baru.
func NewUserHandler(
	userReadRepo user.ReadRepository,
	userWriteRepo user.WriteRepository,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
) *UserHandler {
	return &UserHandler{
		BaseHandler:   NewBaseHandler(cb, qb),
		userReadRepo:  userReadRepo,
		userWriteRepo: userWriteRepo,
	}
}

// RegisterRoutes mendaftarkan semua route untuk manajemen user.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/me", h.GetMe)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List users
// @Tags         users
// @Produce      json
// @Router       /api/v1/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, userSortConfig)

	s, ok := scope.ScopeFromContext(r.Context())
	if !ok {
		h.RespondError(w, http.StatusForbidden, "scope organisasi tidak teridentifikasi")
		return
	}

	users, total, err := h.userReadRepo.List(r.Context(), s, params.Limit, params.Offset)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data user")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"data":   users,
		"total":  total,
		"limit":  params.Limit,
		"offset": params.Offset,
	})
}

// GetByID godoc
// @Summary      Get user by ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	u, err := h.userReadRepo.GetByID(r.Context(), id)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "user tidak ditemukan")
		return
	}

	h.RespondJSON(w, http.StatusOK, u)
}

// GetMe godoc
// @Summary      Get current user profile
// @Tags         users
// @Produce      json
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		h.RespondError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	u, err := h.userReadRepo.GetByID(r.Context(), userID)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "user tidak ditemukan")
		return
	}

	h.RespondJSON(w, http.StatusOK, u)
}

// Create godoc
// @Summary      Create user
// @Tags         users
// @Accept       json
// @Produce      json
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, registercmd.Command{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update user
// @Tags         users
// @Accept       json
// @Param        id   path      string  true  "User ID"
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	existing, err := h.userReadRepo.GetByID(r.Context(), id)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "user tidak ditemukan")
		return
	}

	existing.FullName = req.FullName
	existing.Phone = req.Phone
	existing.Email = req.Email

	if err := h.userWriteRepo.Update(r.Context(), existing); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengupdate user")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete user
// @Tags         users
// @Param        id   path      string  true  "User ID"
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	// Verify user exists before deleting
	_, err = h.userReadRepo.GetByID(r.Context(), id)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "user tidak ditemukan")
		return
	}

	if err := h.userWriteRepo.SoftDelete(r.Context(), id); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal menghapus user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var userSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"full_name":  "full_name",
		"email":      "email",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}
