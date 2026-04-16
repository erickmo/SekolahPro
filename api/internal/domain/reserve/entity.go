// Package reserve mendefinisikan domain untuk Reserve Fund (Dana Cadangan).
//
// Mengelola dana cadangan koperasi sesuai regulasi: pengaturan dana,
// pencatatan transaksi masuk/keluar, dan pelaporan saldo.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package reserve

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── ReserveFund ───────────────────────────────────────────────────────────────

// FundStatus mendefinisikan status dana cadangan.
type FundStatus string

const (
	FundStatusActive   FundStatus = "active"
	FundStatusFrozen   FundStatus = "frozen"
	FundStatusClosed   FundStatus = "closed"
)

// ReserveFund adalah entity utama untuk dana cadangan koperasi.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type ReserveFund struct {
	ID              uuid.UUID  `db:"id"`
	Name            string     `db:"name"`
	FundType        string     `db:"fund_type"`
	Status          FundStatus `db:"status"`
	TargetAmount    int64      `db:"target_amount"`
	CurrentBalance  int64      `db:"current_balance"`
	MinimumBalance  int64      `db:"minimum_balance"`
	ContributionPct float64    `db:"contribution_pct"`
	Description     string     `db:"description"`
	LastCalculatedAt *time.Time `db:"last_calculated_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

// ReserveFundFilter berisi parameter filter opsional untuk list ReserveFund.
type ReserveFundFilter struct {
	FundType *string
	Status   *FundStatus
}

// ReserveFundWriteRepository mendefinisikan operasi write untuk ReserveFund.
type ReserveFundWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *ReserveFund) error
	Update(ctx context.Context, s scope.Scope, e *ReserveFund) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReserveFundReadRepository mendefinisikan operasi read untuk ReserveFund.
type ReserveFundReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*ReserveFund, error)
	List(ctx context.Context, s scope.Scope, filter ReserveFundFilter, limit, offset int, sortBy, order string) ([]*ReserveFund, int, error)
}

// ── ReserveFundTransaction ────────────────────────────────────────────────────

// TransactionType mendefinisikan jenis transaksi dana cadangan.
type TransactionType string

const (
	TransactionTypeContribution   TransactionType = "contribution"
	TransactionTypeWithdrawal     TransactionType = "withdrawal"
	TransactionTypeAdjustment     TransactionType = "adjustment"
	TransactionTypeTransfer       TransactionType = "transfer"
	TransactionTypeInterest       TransactionType = "interest"
)

// ReserveFundTransaction adalah entity untuk transaksi dana cadangan.
type ReserveFundTransaction struct {
	ID              uuid.UUID      `db:"id"`
	ReserveFundID   uuid.UUID      `db:"reserve_fund_id"`
	TransactionType TransactionType `db:"transaction_type"`
	Amount          int64          `db:"amount"`
	BalanceBefore   int64          `db:"balance_before"`
	BalanceAfter    int64          `db:"balance_after"`
	ReferenceNo     string         `db:"reference_no"`
	Description     string         `db:"description"`
	ProcessedBy     uuid.UUID      `db:"processed_by"`
	ProcessedAt     time.Time      `db:"processed_at"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at"`
}

// ReserveFundTransactionFilter berisi parameter filter untuk list transaksi.
type ReserveFundTransactionFilter struct {
	ReserveFundID   *uuid.UUID
	TransactionType *TransactionType
	ProcessedAtFrom *time.Time
	ProcessedAtTo   *time.Time
}

// ReserveFundTransactionWriteRepository mendefinisikan operasi write untuk transaksi.
type ReserveFundTransactionWriteRepository interface {
	SaveTransaction(ctx context.Context, s scope.Scope, e *ReserveFundTransaction) error
	UpdateTransaction(ctx context.Context, s scope.Scope, e *ReserveFundTransaction) error
	DeleteTransaction(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReserveFundTransactionReadRepository mendefinisikan operasi read untuk transaksi.
type ReserveFundTransactionReadRepository interface {
	GetTransactionByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*ReserveFundTransaction, error)
	ListTransactions(ctx context.Context, s scope.Scope, filter ReserveFundTransactionFilter, limit, offset int, sortBy, order string) ([]*ReserveFundTransaction, int, error)
}
