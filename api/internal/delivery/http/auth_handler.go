package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	changepwcmd "github.com/yourorg/boilerplate/internal/command/change_password"
	logincmd "github.com/yourorg/boilerplate/internal/command/login"
	"github.com/yourorg/boilerplate/internal/domain/user"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"

	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// AuthHandler menangani HTTP request untuk autentikasi.
type AuthHandler struct {
	BaseHandler
	loginHandler         *logincmd.Handler
	changePasswordHandler *changepwcmd.Handler
	userReadRepo         user.ReadRepository
	jwtSvc               *jwtpkg.Service
}

// NewAuthHandler membuat instance AuthHandler baru.
func NewAuthHandler(
	loginHandler *logincmd.Handler,
	changePasswordHandler *changepwcmd.Handler,
	userReadRepo user.ReadRepository,
	jwtSvc *jwtpkg.Service,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
) *AuthHandler {
	return &AuthHandler{
		BaseHandler:           NewBaseHandler(cb, qb),
		loginHandler:         loginHandler,
		changePasswordHandler: changePasswordHandler,
		userReadRepo:         userReadRepo,
		jwtSvc:               jwtSvc,
	}
}

// RegisterRoutes mendaftarkan semua route untuk autentikasi.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/change-password", h.ChangePassword)
	})
}

// Login godoc
// @Summary      Login user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  logincmd.Result
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	result, err := h.loginHandler.Handle(r.Context(), logincmd.Command{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case user.ErrEmailEmpty, user.ErrPasswordEmpty:
			h.RespondError(w, http.StatusBadRequest, err.Error())
		case user.ErrWrongPassword, user.ErrInactive, user.ErrNotFound:
			h.RespondError(w, http.StatusUnauthorized, "email atau password salah")
		default:
			h.RespondError(w, http.StatusInternalServerError, "gagal login")
		}
		return
	}

	h.RespondJSON(w, http.StatusOK, result)
}

// Refresh godoc
// @Summary      Refresh token
// @Tags         auth
// @Produce      json
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
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

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	pair, err := h.jwtSvc.GenerateTokenPair(userID, claims.Email, claims.Role, tenantID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal generate token")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         auth
// @Accept       json
// @Router       /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		h.RespondError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	if err := h.changePasswordHandler.Handle(r.Context(), changepwcmd.Command{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		switch err {
		case user.ErrPasswordEmpty, user.ErrPasswordWeak:
			h.RespondError(w, http.StatusBadRequest, err.Error())
		case user.ErrWrongPassword:
			h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			h.RespondError(w, http.StatusInternalServerError, "gagal mengubah password")
		}
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "password berhasil diubah"})
}
