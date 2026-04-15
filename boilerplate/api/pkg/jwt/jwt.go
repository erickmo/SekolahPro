// Package jwt menyediakan utility untuk generate dan validasi JWT token.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrTokenInvalid dikembalikan ketika token tidak valid atau sudah expired.
var ErrTokenInvalid = errors.New("token invalid or expired")

// Claims berisi payload JWT dengan hierarki organisasi 4 level.
//
// Dua fase token:
//   - Fase 1 (setelah login): TenantID ada, CompanyID/BranchID/WarehouseID kosong.
//   - Fase 2 (setelah select-scope): semua level yang relevan terisi.
type Claims struct {
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	TenantID    string  `json:"tenant_id"`              // Selalu ada — L1
	CompanyID   *string `json:"company_id,omitempty"`   // Terisi setelah select-scope — L2
	BranchID    *string `json:"branch_id,omitempty"`    // Opsional — L3
	WarehouseID *string `json:"warehouse_id,omitempty"` // Opsional — L4
	jwt.RegisteredClaims
}

// GetTenantID mengimplementasi scope.scopeClaimsReader.
func (c *Claims) GetTenantID() string { return c.TenantID }

// GetCompanyID mengimplementasi scope.scopeClaimsReader dan tenant.claimsReader.
func (c *Claims) GetCompanyID() *string { return c.CompanyID }

// GetBranchID mengimplementasi scope.scopeClaimsReader.
func (c *Claims) GetBranchID() *string { return c.BranchID }

// GetWarehouseID mengimplementasi scope.scopeClaimsReader.
func (c *Claims) GetWarehouseID() *string { return c.WarehouseID }

// TokenPair berisi access token dan refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Service menangani operasi JWT.
type Service struct {
	secret            []byte
	expiryHours       int
	issuer            string
	refreshMultiplier int
}

// NewService membuat JWT service baru.
func NewService(secret string, expiryHours int, issuer string, refreshMultiplier int) *Service {
	return &Service{
		secret:            []byte(secret),
		expiryHours:       expiryHours,
		issuer:            issuer,
		refreshMultiplier: refreshMultiplier,
	}
}

// GenerateTokenPair membuat token fase 1 — hanya berisi TenantID, tanpa scope perusahaan.
// Dipakai langsung setelah login, sebelum pengguna memilih perusahaan aktif.
func (s *Service) GenerateTokenPair(userID uuid.UUID, email, role string, tenantID uuid.UUID) (*TokenPair, error) {
	claims := &Claims{
		UserID:   userID.String(),
		Email:    email,
		Role:     role,
		TenantID: tenantID.String(),
	}
	return s.signTokenPair(claims)
}

// GenerateScopedTokenPair membuat token fase 2 — berisi TenantID + scope organisasi lengkap.
// Dipakai setelah pengguna memilih perusahaan (dan opsional cabang/gudang).
func (s *Service) GenerateScopedTokenPair(
	userID uuid.UUID,
	email, role string,
	tenantID uuid.UUID,
	companyID uuid.UUID,
	branchID *uuid.UUID,
	warehouseID *uuid.UUID,
) (*TokenPair, error) {
	companyStr := companyID.String()
	claims := &Claims{
		UserID:    userID.String(),
		Email:     email,
		Role:      role,
		TenantID:  tenantID.String(),
		CompanyID: &companyStr,
	}

	if branchID != nil && *branchID != uuid.Nil {
		str := branchID.String()
		claims.BranchID = &str
	}
	if warehouseID != nil && *warehouseID != uuid.Nil {
		str := warehouseID.String()
		claims.WarehouseID = &str
	}

	return s.signTokenPair(claims)
}

// signTokenPair menandatangani claims menjadi access + refresh token.
func (s *Service) signTokenPair(claims *Claims) (*TokenPair, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(s.expiryHours) * time.Hour)
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.UserID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:    s.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := *claims
	refreshClaims.ExpiresAt = jwt.NewNumericDate(
		time.Now().UTC().Add(time.Duration(s.expiryHours*s.refreshMultiplier) * time.Hour),
	)

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// ValidateClaims memvalidasi token dan mengembalikan claims-nya.
func (s *Service) ValidateClaims(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
