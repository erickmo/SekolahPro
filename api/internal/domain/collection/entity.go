// Package collection mendefinisikan domain untuk manajemen penagihan pinjaman (loan collection).
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package collection

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Aging Bucket ─────────────────────────────────────────────────────────────

// AgingBucket merepresentasikan klasifikasi keterlambatan berdasarkan DPD (Days Past Due).
type AgingBucket string

const (
	AgingBucketCurrent   AgingBucket = "current"
	AgingBucketDPW1To7   AgingBucket = "dpw_1_7"
	AgingBucketDPW8To30  AgingBucket = "dpw_8_30"
	AgingBucketDPW31To60 AgingBucket = "dpw_31_60"
	AgingBucketDPW61To90 AgingBucket = "dpw_61_90"
	AgingBucketDPW91To180 AgingBucket = "dpw_91_180"
	AgingBucketDPW181To270 AgingBucket = "dpw_181_270"
	AgingBucketDPW271To360 AgingBucket = "dpw_271_360"
	AgingBucketDPW360Plus AgingBucket = "dpw_360_plus"
)

// ── Case Status ──────────────────────────────────────────────────────────────

// CaseStatus merepresentasikan status collection case.
type CaseStatus string

const (
	CaseStatusActive      CaseStatus = "active"
	CaseStatusRestructured CaseStatus = "restructured"
	CaseStatusLegal       CaseStatus = "legal"
	CaseStatusWrittenOff  CaseStatus = "written_off"
	CaseStatusSettled     CaseStatus = "settled"
	CaseStatusClosed      CaseStatus = "closed"
)

// ── Resolution Type ──────────────────────────────────────────────────────────

// ResolutionType merepresentasikan cara penyelesaian collection case.
type ResolutionType string

const (
	ResolutionFullPayment          ResolutionType = "full_payment"
	ResolutionRestructuring        ResolutionType = "restructuring"
	ResolutionCollateralLiquidation ResolutionType = "collateral_liquidation"
	ResolutionGuaranteeClaim       ResolutionType = "guarantee_claim"
	ResolutionWriteOff             ResolutionType = "write_off"
	ResolutionPartialSettlement    ResolutionType = "partial_settlement"
)

// ── Activity Type ────────────────────────────────────────────────────────────

// ActivityType merepresentasikan jenis aktivitas penagihan.
type ActivityType string

const (
	ActivityAutoReminder         ActivityType = "auto_reminder"
	ActivityPhoneCall            ActivityType = "phone_call"
	ActivityHomeVisit            ActivityType = "home_visit"
	ActivitySPLetter             ActivityType = "_letter"
	ActivityNegotiation          ActivityType = "negotiation"
	ActivityRestructuringOffer   ActivityType = "restructuring_offer"
	ActivityLegalNotice          ActivityType = "legal_notice"
	ActivityCollateralProcessing ActivityType = "collateral_processing"
	ActivityGuaranteeContact     ActivityType = "guarantee_contact"
	ActivityPaymentReceived      ActivityType = "payment_received"
)

// ── Contact Result ───────────────────────────────────────────────────────────

// ContactResult merepresentasikan hasil upaya kontak terhadap nasabah.
type ContactResult string

const (
	ContactResultContacted      ContactResult = "contacted"
	ContactResultNoAnswer       ContactResult = "no_answer"
	ContactResultBusy           ContactResult = "busy"
	ContactResultWrongNumber    ContactResult = "wrong_number"
	ContactResultMessageLeft    ContactResult = "message_left"
	ContactResultVisitedMet     ContactResult = "visited_met"
	ContactResultVisitedNotHome ContactResult = "visited_not_home"
	ContactResultLetterSent     ContactResult = "letter_sent"
	ContactResultLetterDelivered ContactResult = "letter_delivered"
	ContactResultLetterReturned ContactResult = "letter_returned"
)

// ── Nasabah Response ─────────────────────────────────────────────────────────

// NasabahResponse merepresentasikan respons nasabah terhadap upaya penagihan.
type NasabahResponse string

const (
	NasabahResponseWillingToPay     NasabahResponse = "willing_to_pay"
	NasabahResponseNeedsRestructuring NasabahResponse = "needs_restructuring"
	NasabahResponseRefusesToPay     NasabahResponse = "refuses_to_pay"
	NasabahResponseUnableToPay      NasabahResponse = "unable_to_pay"
	NasabahResponseDispute          NasabahResponse = "dispute"
	NasabahResponsePromiseToPay     NasabahResponse = "promise_to_pay"
	NasabahResponseNoResponse       NasabahResponse = "no_response"
)

// ── CollectionCase Entity ────────────────────────────────────────────────────

// CollectionCase adalah entity utama untuk tracking penagihan pinjaman.
type CollectionCase struct {
	ID                        uuid.UUID      `db:"id"`
	CaseNumber                string         `db:"case_number"`
	CurrentDPD                int            `db:"current_dpd"`
	AgingBucket               AgingBucket    `db:"aging_bucket"`
	TotalOverdueAmount        int64          `db:"total_overdue_amount"`
	TotalOverdueInstallments  int            `db:"total_overdue_installments"`
	PinjamanID                uuid.UUID      `db:"pinjaman_id"`
	NasabahID                 uuid.UUID      `db:"nasabah_id"`
	AssignedCollectorID       *uuid.UUID     `db:"assigned_collector_id"`
	AssignedAt                *time.Time     `db:"assigned_at"`
	EscalationLevel           int            `db:"escalation_level"`
	Status                    CaseStatus     `db:"status"`
	ResolutionType            *ResolutionType `db:"resolution_type"`
	OpenedAt                  time.Time      `db:"opened_at"`
	ClosedAt                  *time.Time     `db:"closed_at"`
	CreatedAt                 time.Time      `db:"created_at"`
	UpdatedAt                 time.Time      `db:"updated_at"`
	DeletedAt                 *time.Time     `db:"deleted_at"`
}

// ── Activity Entity ──────────────────────────────────────────────────────────

// Activity merepresentasikan satu aktivitas penagihan pada sebuah collection case.
type Activity struct {
	ID               uuid.UUID       `db:"id"`
	CaseID           uuid.UUID       `db:"case_id"`
	ActivityType     ActivityType    `db:"activity_type"`
	ActivityDate     time.Time       `db:"activity_date"`
	PerformedBy      uuid.UUID       `db:"performed_by"`
	ContactResult    ContactResult   `db:"contact_result"`
	Notes            string          `db:"notes"`
	NasabahResponse  *NasabahResponse `db:"nasabah_response"`
	PromiseAmount    *int64          `db:"promise_amount"`
	PromiseDate      *time.Time      `db:"promise_date"`
	DocumentIDs      []uuid.UUID     `db:"document_ids"`
	FollowupRequired bool            `db:"followup_required"`
	FollowupDate     *time.Time      `db:"followup_date"`
	FollowupType     *string         `db:"followup_type"`
	CreatedAt        time.Time       `db:"created_at"`
}

// ── Performance Entity ───────────────────────────────────────────────────────

// Performance merepresentasikan statistik performa collector dalam satu periode.
type Performance struct {
	ID                   uuid.UUID `db:"id"`
	CollectorID          uuid.UUID `db:"collector_id"`
	PeriodMonth          string    `db:"period_month"`
	TotalCalls           int       `db:"total_calls"`
	SuccessfulContacts   int       `db:"successful_contacts"`
	TotalVisits          int       `db:"total_visits"`
	SuccessfulVisits     int       `db:"successful_visits"`
	CasesHandled         int       `db:"cases_handled"`
	CasesResolved        int       `db:"cases_resolved"`
	TotalAmountCollected int64     `db:"total_amount_collected"`
	PromiseToPayCount    int       `db:"promise_to_pay_count"`
	PromiseKeptCount     int       `db:"promise_kept_count"`
	ContactRatePct       float64   `db:"contact_rate_pct"`
	ResolutionRatePct    float64   `db:"resolution_rate_pct"`
	CollectionRatePct    float64   `db:"collection_rate_pct"`
	PromiseKeptRatePct   float64   `db:"promise_kept_rate_pct"`
	CalculatedAt         time.Time `db:"calculated_at"`
	CreatedAt            time.Time `db:"created_at"`
}

// ── CaseFilter untuk query list ──────────────────────────────────────────────

// CaseFilter berisi parameter filter untuk listing collection cases.
type CaseFilter struct {
	Status      *CaseStatus
	AgingBucket *AgingBucket
	CollectorID *uuid.UUID
	NasabahID   *uuid.UUID
}

// ── Repository Interfaces ────────────────────────────────────────────────────

// CaseWriteRepository mendefinisikan operasi write untuk CollectionCase.
type CaseWriteRepository interface {
	SaveCase(ctx context.Context, s scope.Scope, c *CollectionCase) error
	UpdateCase(ctx context.Context, s scope.Scope, c *CollectionCase) error
	DeleteCase(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// CaseReadRepository mendefinisikan operasi read untuk CollectionCase.
type CaseReadRepository interface {
	GetCaseByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*CollectionCase, error)
	ListCases(ctx context.Context, s scope.Scope, filter CaseFilter, limit, offset int, sortBy, order string) ([]*CollectionCase, int, error)
}

// ActivityWriteRepository mendefinisikan operasi write untuk Activity.
type ActivityWriteRepository interface {
	SaveActivity(ctx context.Context, s scope.Scope, a *Activity) error
}

// ActivityReadRepository mendefinisikan operasi read untuk Activity.
type ActivityReadRepository interface {
	ListActivities(ctx context.Context, s scope.Scope, caseID uuid.UUID, activityType *ActivityType, limit, offset int) ([]*Activity, int, error)
}

// PerformanceWriteRepository mendefinisikan operasi write untuk Performance.
type PerformanceWriteRepository interface {
	SavePerformance(ctx context.Context, s scope.Scope, p *Performance) error
}

// PerformanceReadRepository mendefinisikan operasi read untuk Performance.
type PerformanceReadRepository interface {
	ListPerformances(ctx context.Context, s scope.Scope, collectorID *uuid.UUID, periodMonth string, limit, offset int) ([]*Performance, int, error)
}

// CalculateAgingBucket mengembalikan AgingBucket berdasarkan DPD (Days Past Due).
func CalculateAgingBucket(dpd int) AgingBucket {
	switch {
	case dpd <= 0:
		return AgingBucketCurrent
	case dpd <= 7:
		return AgingBucketDPW1To7
	case dpd <= 30:
		return AgingBucketDPW8To30
	case dpd <= 60:
		return AgingBucketDPW31To60
	case dpd <= 90:
		return AgingBucketDPW61To90
	case dpd <= 180:
		return AgingBucketDPW91To180
	case dpd <= 270:
		return AgingBucketDPW181To270
	case dpd <= 360:
		return AgingBucketDPW271To360
	default:
		return AgingBucketDPW360Plus
	}
}
