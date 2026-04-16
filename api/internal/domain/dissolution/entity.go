// Package dissolution mendefinisikan domain untuk proses pembubaran koperasi.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package dissolution

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// DissolutionType menentukan jenis pembubaran koperasi.
type DissolutionType string

const (
	DissolutionTypeVoluntary      DissolutionType = "voluntary"
	DissolutionTypeGovernmentOrder DissolutionType = "government_order"
	DissolutionTypeMerger         DissolutionType = "merger"
	DissolutionTypeSplit          DissolutionType = "split"
	DissolutionTypeBankruptcy     DissolutionType = "bankruptcy"
)

// ValidDissolutionTypes berisi semua jenis pembubaran yang valid.
var ValidDissolutionTypes = map[DissolutionType]bool{
	DissolutionTypeVoluntary:       true,
	DissolutionTypeGovernmentOrder: true,
	DissolutionTypeMerger:          true,
	DissolutionTypeSplit:           true,
	DissolutionTypeBankruptcy:      true,
}

// Stage menentukan tahapan proses pembubaran.
type Stage string

const (
	StageAnnounced         Stage = "announced"
	StageClaimPeriod       Stage = "claim_period"
	StageAssetLiquidation  Stage = "asset_liquidation"
	StageDebtSettlement    Stage = "debt_settlement"
	StageMemberDistribution Stage = "member_distribution"
	StageFinalReport       Stage = "final_report"
	StageClosed            Stage = "closed"
)

// StageOrder menentukan urutan transisi tahapan yang valid.
var StageOrder = []Stage{
	StageAnnounced,
	StageClaimPeriod,
	StageAssetLiquidation,
	StageDebtSettlement,
	StageMemberDistribution,
	StageFinalReport,
	StageClosed,
}

// NextStage mengembalikan stage berikutnya dari stage saat ini.
// Mengembalikan (stage, true) jika transisi valid, atau ("", false) jika sudah di akhir.
func NextStage(current Stage) (Stage, bool) {
	for i, s := range StageOrder {
		if s == current && i+1 < len(StageOrder) {
			return StageOrder[i+1], true
		}
	}
	return "", false
}

// Status menentukan status proses pembubaran.
type Status string

const (
	StatusInitiated   Status = "initiated"
	StatusInProgress  Status = "in_progress"
	StatusCompleted   Status = "completed"
	StatusCancelled   Status = "cancelled"
)

// DissolutionProcess adalah entity utama untuk proses pembubaran koperasi.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type DissolutionProcess struct {
	ID                    uuid.UUID      `db:"id"`
	DissolutionType       DissolutionType `db:"dissolution_type"`
	RatMeetingID          *uuid.UUID     `db:"rat_meeting_id"`
	Reason                string         `db:"reason"`
	EffectiveDate         time.Time      `db:"effective_date"`
	LiquidatorIDs         []uuid.UUID    `db:"liquidator_ids"`
	SupervisorID          *uuid.UUID     `db:"supervisor_id"`
	Stage                 Stage          `db:"stage"`
	ClaimDeadline         *time.Time     `db:"claim_deadline"`
	TotalAssets           *int64         `db:"total_assets"`
	TotalLiabilities      *int64         `db:"total_liabilities"`
	NetEquity             *int64         `db:"net_equity"`
	DistributionPerMember *int64         `db:"distribution_per_member"`
	Status                Status         `db:"status"`
	FinalReportDocID      *uuid.UUID     `db:"final_report_doc_id"`
	ClosedAt              *time.Time     `db:"closed_at"`
	InitiatedAt           time.Time      `db:"initiated_at"`
	InitiatedBy           uuid.UUID      `db:"initiated_by"`
	DeletedAt             *time.Time     `db:"deleted_at"`
	CreatedAt             time.Time      `db:"created_at"`
}

// ListFilter mendefinisikan filter untuk query list dissolution processes.
type ListFilter struct {
	DissolutionType *DissolutionType
	Status          *Status
	Stage           *Stage
}

// WriteRepository mendefinisikan operasi write untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *DissolutionProcess) error
	Update(ctx context.Context, s scope.Scope, e *DissolutionProcess) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*DissolutionProcess, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*DissolutionProcess, int, error)
}
