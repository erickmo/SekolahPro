package branch_report

import "github.com/google/uuid"

// BranchReportGeneratedEvent dipublikasikan saat laporan keuangan cabang berhasil dibuat.
type BranchReportGeneratedEvent struct {
	TenantID  uuid.UUID  `json:"tenant_id"`
	CompanyID uuid.UUID  `json:"company_id"`
	ID        uuid.UUID  `json:"id"`
	BranchID  *uuid.UUID `json:"branch_id"`
	PeriodType PeriodType `json:"period_type"`
}

func (e BranchReportGeneratedEvent) EventName() string   { return "branch_report.generated" }
func (e BranchReportGeneratedEvent) AggregateID() string { return e.ID.String() }

// BranchReportApprovedEvent dipublikasikan saat laporan keuangan cabang disetujui.
type BranchReportApprovedEvent struct {
	TenantID   uuid.UUID  `json:"tenant_id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	ID         uuid.UUID  `json:"id"`
	BranchID   *uuid.UUID `json:"branch_id"`
	ApprovedBy uuid.UUID  `json:"approved_by"`
}

func (e BranchReportApprovedEvent) EventName() string   { return "branch_report.approved" }
func (e BranchReportApprovedEvent) AggregateID() string { return e.ID.String() }

// ConsolidatedReportGeneratedEvent dipublikasikan saat laporan konsolidasi berhasil dibuat.
type ConsolidatedReportGeneratedEvent struct {
	TenantID   uuid.UUID  `json:"tenant_id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	PeriodType PeriodType `json:"period_type"`
	BranchCount int       `json:"branch_count"`
}

func (e ConsolidatedReportGeneratedEvent) EventName() string { return "branch_report.consolidated" }
func (e ConsolidatedReportGeneratedEvent) AggregateID() string { return "consolidated" }
