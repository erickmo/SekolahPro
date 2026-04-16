// Package insurance mendefinisikan domain untuk asuransi (Products, Policies, Claims).
//
// Mengelola produk asuransi, polis anggota, dan proses klaim
// untuk perlindungan anggota koperasi sekolah.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package insurance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Product ───────────────────────────────────────────────────────────────────

// ProductType mendefinisikan jenis produk asuransi.
type ProductType string

const (
	ProductTypeLife      ProductType = "life"
	ProductTypeHealth    ProductType = "health"
	ProductTypeAccident  ProductType = "accident"
	ProductTypeEducation ProductType = "education"
	ProductTypeProperty  ProductType = "property"
)

// ProductStatus mendefinisikan status produk.
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusActive    ProductStatus = "active"
	ProductStatusInactive  ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

// Product adalah entity untuk produk asuransi.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type Product struct {
	ID              uuid.UUID    `db:"id"`
	Name            string       `db:"name"`
	Code            string       `db:"code"`
	ProductType     ProductType  `db:"product_type"`
	Status          ProductStatus `db:"status"`
	ProviderName    string       `db:"provider_name"`
	PremiumAmount   int64        `db:"premium_amount"`
	CoverageAmount  int64        `db:"coverage_amount"`
	PremiumFrequency string      `db:"premium_frequency"`
	TermMonths      int          `db:"term_months"`
	Description     string       `db:"description"`
	TermsConditions string       `db:"terms_conditions"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
	DeletedAt       *time.Time   `db:"deleted_at"`
}

// ProductFilter berisi parameter filter opsional untuk list Product.
type ProductFilter struct {
	ProductType *ProductType
	Status      *ProductStatus
}

// ProductWriteRepository mendefinisikan operasi write untuk Product.
type ProductWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *Product) error
	Update(ctx context.Context, s scope.Scope, e *Product) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ProductReadRepository mendefinisikan operasi read untuk Product.
type ProductReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Product, error)
	List(ctx context.Context, s scope.Scope, filter ProductFilter, limit, offset int, sortBy, order string) ([]*Product, int, error)
}

// ── Policy ────────────────────────────────────────────────────────────────────

// PolicyStatus mendefinisikan status polis.
type PolicyStatus string

const (
	PolicyStatusPending    PolicyStatus = "pending"
	PolicyStatusActive     PolicyStatus = "active"
	PolicyStatusLapsed     PolicyStatus = "lapsed"
	PolicyStatusCancelled  PolicyStatus = "cancelled"
	PolicyStatusExpired    PolicyStatus = "expired"
	PolicyStatusClaimed    PolicyStatus = "claimed"
)

// Policy adalah entity untuk polis asuransi anggota.
type Policy struct {
	ID              uuid.UUID    `db:"id"`
	ProductID       uuid.UUID    `db:"product_id"`
	NasabahID       uuid.UUID    `db:"nasabah_id"`
	PolicyNo        string       `db:"policy_no"`
	Status          PolicyStatus `db:"status"`
	StartDate       time.Time    `db:"start_date"`
	EndDate         time.Time    `db:"end_date"`
	PremiumAmount   int64        `db:"premium_amount"`
	CoverageAmount  int64        `db:"coverage_amount"`
	LastPremiumPaid *time.Time   `db:"last_premium_paid"`
	NextPremiumDue  *time.Time   `db:"next_premium_due"`
	BeneficiaryName string       `db:"beneficiary_name"`
	BeneficiaryRelation string    `db:"beneficiary_relation"`
	IssuedAt        *time.Time   `db:"issued_at"`
	IssuedBy        *uuid.UUID   `db:"issued_by"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
	DeletedAt       *time.Time   `db:"deleted_at"`
}

// PolicyFilter berisi parameter filter opsional untuk list Policy.
type PolicyFilter struct {
	ProductID  *uuid.UUID
	NasabahID  *uuid.UUID
	Status     *PolicyStatus
}

// PolicyWriteRepository mendefinisikan operasi write untuk Policy.
type PolicyWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *Policy) error
	Update(ctx context.Context, s scope.Scope, e *Policy) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// PolicyReadRepository mendefinisikan operasi read untuk Policy.
type PolicyReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Policy, error)
	List(ctx context.Context, s scope.Scope, filter PolicyFilter, limit, offset int, sortBy, order string) ([]*Policy, int, error)
}

// ── Claim ─────────────────────────────────────────────────────────────────────

// ClaimType mendefinisikan jenis klaim.
type ClaimType string

const (
	ClaimTypeDeath         ClaimType = "death"
	ClaimTypeHospitalization ClaimType = "hospitalization"
	ClaimTypeAccident      ClaimType = "accident"
	ClaimTypeDisability    ClaimType = "disability"
	ClaimTypeCriticalIllness ClaimType = "critical_illness"
)

// ClaimStatus mendefinisikan status klaim.
type ClaimStatus string

const (
	ClaimStatusSubmitted  ClaimStatus = "submitted"
	ClaimStatusUnderReview ClaimStatus = "under_review"
	ClaimStatusApproved   ClaimStatus = "approved"
	ClaimStatusRejected   ClaimStatus = "rejected"
	ClaimStatusPaid       ClaimStatus = "paid"
)

// Claim adalah entity untuk klaim asuransi.
type Claim struct {
	ID              uuid.UUID    `db:"id"`
	PolicyID        uuid.UUID    `db:"policy_id"`
	ClaimType       ClaimType    `db:"claim_type"`
	Status          ClaimStatus  `db:"status"`
	ClaimAmount     int64        `db:"claim_amount"`
	ApprovedAmount  *int64       `db:"approved_amount"`
	IncidentDate    time.Time    `db:"incident_date"`
	SubmittedAt     time.Time    `db:"submitted_at"`
	SubmittedBy     uuid.UUID    `db:"submitted_by"`
	Description     string       `db:"description"`
	DocumentIDs     []uuid.UUID  `db:"document_ids"`
	ReviewedAt      *time.Time   `db:"reviewed_at"`
	ReviewedBy      *uuid.UUID   `db:"reviewed_by"`
	RejectionReason string       `db:"rejection_reason"`
	PaidAt          *time.Time   `db:"paid_at"`
	PaidBy          *uuid.UUID   `db:"paid_by"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
	DeletedAt       *time.Time   `db:"deleted_at"`
}

// ClaimFilter berisi parameter filter opsional untuk list Claim.
type ClaimFilter struct {
	PolicyID  *uuid.UUID
	ClaimType *ClaimType
	Status    *ClaimStatus
}

// ClaimWriteRepository mendefinisikan operasi write untuk Claim.
type ClaimWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *Claim) error
	Update(ctx context.Context, s scope.Scope, e *Claim) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ClaimReadRepository mendefinisikan operasi read untuk Claim.
type ClaimReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Claim, error)
	List(ctx context.Context, s scope.Scope, filter ClaimFilter, limit, offset int, sortBy, order string) ([]*Claim, int, error)
}
