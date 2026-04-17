// Package laporan_regulasi adalah domain Vernon untuk pelaporan regulasi koperasi.
//
// Laporan regulasi mencakup pelaporan ke Dinas Koperasi, OJK, laporan syariah,
// dan laporan internal manajemen. Domain ini memiliki 3 tabel Vernon:
// laporan_config (konfigurasi), laporan (record), laporan_versi (versioning).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package laporan_regulasi

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// laporan_config — Field name constants
// ============================================================================

const (
	CfgFieldReportType             = "report_type"
	CfgFieldReportName             = "report_name"
	CfgFieldReportCategory         = "report_category"
	CfgFieldDescription            = "description"
	CfgFieldFrequency              = "frequency"
	CfgFieldTemplateConfig         = "template_config"
	CfgFieldIsActive               = "is_active"
	CfgFieldDeadlineDaysAfterPeriod = "deadline_days_after_period"
)

// laporan_config — Report category constants.
const (
	CategoryDinasKoperasi = "dinas_koperasi"
	CategoryOJK           = "ojk"
	CategoryIslamic       = "islamic"
	CategoryInternal      = "internal"
)

// laporan_config — Frequency constants.
const (
	FrequencyDaily     = "daily"
	FrequencyWeekly    = "weekly"
	FrequencyMonthly   = "monthly"
	FrequencyQuarterly = "quarterly"
	FrequencyAnnually  = "annually"
)

var validReportCategories = map[string]bool{
	CategoryDinasKoperasi: true, CategoryOJK: true,
	CategoryIslamic: true, CategoryInternal: true,
}

var validFrequencies = map[string]bool{
	FrequencyDaily: true, FrequencyWeekly: true,
	FrequencyMonthly: true, FrequencyQuarterly: true,
	FrequencyAnnually: true,
}

// ConfigDescriptor mengimplementasi vernon.DomainDescriptor untuk laporan_config.
type ConfigDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ConfigDescriptor) TableName() string { return "laporan_config" }

// DefaultRels — laporan_config tidak memiliki BelongsTo.
func (d *ConfigDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant laporan_config sebelum write.
func (d *ConfigDescriptor) Validate(data map[string]any) error {
	return validateConfigRequired(data)
}

// validateConfigRequired memeriksa field wajib laporan_config.
func validateConfigRequired(data map[string]any) error {
	reportType, _ := data[CfgFieldReportType].(string)
	if reportType == "" {
		return errors.New("report_type wajib diisi")
	}

	reportName, _ := data[CfgFieldReportName].(string)
	if reportName == "" {
		return errors.New("report_name wajib diisi")
	}

	category, _ := data[CfgFieldReportCategory].(string)
	if category == "" {
		return errors.New("report_category wajib diisi")
	}
	if !validReportCategories[category] {
		return fmt.Errorf("report_category tidak valid: %q", category)
	}

	freq, _ := data[CfgFieldFrequency].(string)
	if freq == "" {
		return errors.New("frequency wajib diisi")
	}
	if !validFrequencies[freq] {
		return fmt.Errorf("frequency tidak valid: %q", freq)
	}
	return nil
}

// ============================================================================
// laporan — Field name constants
// ============================================================================

const (
	LapFieldLaporanConfigID     = "laporan_config_id"
	LapFieldReportType          = "report_type"
	LapFieldReportCategory      = "report_category"
	LapFieldPeriodType          = "period_type"
	LapFieldPeriodStart         = "period_start"
	LapFieldPeriodEnd           = "period_end"
	LapFieldPeriodLabel         = "period_label"
	LapFieldReportData          = "report_data"
	LapFieldGeneratedAt         = "generated_at"
	LapFieldGeneratedBy         = "generated_by"
	LapFieldBranchID            = "branch_id"
	LapFieldRevisionOf          = "revision_of"
	LapFieldSubmittedAt         = "submitted_at"
	LapFieldSubmittedBy         = "submitted_by"
	LapFieldReviewedAt          = "reviewed_at"
	LapFieldReviewedBy          = "reviewed_by"
	LapFieldAnomalyFlags        = "anomaly_flags"
	LapFieldStatus              = "status"
	LapFieldConsolidationLevel  = "consolidation_level"
)

// laporan — Status constants.
const (
	LapStatusDraft     = "draft"
	LapStatusReviewed  = "reviewed"
	LapStatusFinal     = "final"
	LapStatusSubmitted = "submitted"
	LapStatusOverdue   = "overdue"
)

// laporan — Consolidation level constants.
const (
	ConsolidationBranch  = "branch"
	ConsolidationCompany = "company"
	ConsolidationTenant  = "tenant"
)

// laporan — Relation name constants.
const (
	RelBranch        = "branch"
	RelLaporanConfig = "laporan_config"
)

var validLaporanStatuses = map[string]bool{
	LapStatusDraft: true, LapStatusReviewed: true,
	LapStatusFinal: true, LapStatusSubmitted: true,
	LapStatusOverdue: true,
}

var validPeriodTypes = map[string]bool{
	FrequencyDaily: true, FrequencyWeekly: true,
	FrequencyMonthly: true, FrequencyQuarterly: true,
	FrequencyAnnually: true,
}

var validConsolidationLevels = map[string]bool{
	ConsolidationBranch: true, ConsolidationCompany: true,
	ConsolidationTenant: true,
}

// LaporanDescriptor mengimplementasi vernon.DomainDescriptor untuk laporan.
type LaporanDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LaporanDescriptor) TableName() string { return "laporan" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo branches dan laporan_config (autoload).
func (d *LaporanDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelBranch: {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         LapFieldBranchID,
			LocalKey:   LapFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RelLaporanConfig: {
			Domain:     "laporan_config",
			Type:       vernon.RelBelongsTo,
			FK:         LapFieldLaporanConfigID,
			LocalKey:   LapFieldLaporanConfigID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"report_type", "report_name", "report_category"},
		},
	}
}

// Validate memvalidasi invariant laporan sebelum write.
func (d *LaporanDescriptor) Validate(data map[string]any) error {
	if err := validateLaporanRequired(data); err != nil {
		return err
	}
	return validateLaporanEnums(data)
}

// validateLaporanRequired memeriksa field wajib laporan.
func validateLaporanRequired(data map[string]any) error {
	configID, _ := data[LapFieldLaporanConfigID].(string)
	if configID == "" {
		return errors.New("laporan_config_id wajib diisi")
	}

	reportType, _ := data[LapFieldReportType].(string)
	if reportType == "" {
		return errors.New("report_type wajib diisi")
	}

	category, _ := data[LapFieldReportCategory].(string)
	if category == "" {
		return errors.New("report_category wajib diisi")
	}

	periodType, _ := data[LapFieldPeriodType].(string)
	if periodType == "" {
		return errors.New("period_type wajib diisi")
	}

	periodStart, _ := data[LapFieldPeriodStart].(string)
	if periodStart == "" {
		return errors.New("period_start wajib diisi")
	}

	periodEnd, _ := data[LapFieldPeriodEnd].(string)
	if periodEnd == "" {
		return errors.New("period_end wajib diisi")
	}

	periodLabel, _ := data[LapFieldPeriodLabel].(string)
	if periodLabel == "" {
		return errors.New("period_label wajib diisi")
	}

	reportData := data[LapFieldReportData]
	if reportData == nil {
		return errors.New("report_data wajib diisi")
	}

	generatedAt, _ := data[LapFieldGeneratedAt].(string)
	if generatedAt == "" {
		return errors.New("generated_at wajib diisi")
	}

	generatedBy, _ := data[LapFieldGeneratedBy].(string)
	if generatedBy == "" {
		return errors.New("generated_by wajib diisi")
	}
	return nil
}

// validateLaporanEnums memeriksa enum fields laporan.
func validateLaporanEnums(data map[string]any) error {
	category, _ := data[LapFieldReportCategory].(string)
	if category != "" && !validReportCategories[category] {
		return fmt.Errorf("report_category tidak valid: %q", category)
	}

	periodType, _ := data[LapFieldPeriodType].(string)
	if periodType != "" && !validPeriodTypes[periodType] {
		return fmt.Errorf("period_type tidak valid: %q", periodType)
	}

	status, _ := data[LapFieldStatus].(string)
	if status != "" && !validLaporanStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	level, _ := data[LapFieldConsolidationLevel].(string)
	if level != "" && !validConsolidationLevels[level] {
		return fmt.Errorf("consolidation_level tidak valid: %q", level)
	}
	return nil
}

// ============================================================================
// laporan_versi — Field name constants
// ============================================================================

const (
	VersiFieldLaporanID    = "laporan_id"
	VersiFieldVersionNo    = "version_number"
	VersiFieldReportData   = "report_data"
	VersiFieldCreatedReason = "created_reason"
	VersiFieldChangeSummary = "change_summary"
)

// laporan_versi — Created reason constants.
const (
	ReasonAutoGenerate  = "auto_generate"
	ReasonManualTrigger = "manual_trigger"
	ReasonRegenerate    = "regenerate"
	ReasonRevision      = "revision"
)

// laporan_versi — Relation name constants.
const (
	RelLaporan = "laporan"
)

var validCreatedReasons = map[string]bool{
	ReasonAutoGenerate: true, ReasonManualTrigger: true,
	ReasonRegenerate: true, ReasonRevision: true,
}

// VersiDescriptor mengimplementasi vernon.DomainDescriptor untuk laporan_versi.
type VersiDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *VersiDescriptor) TableName() string { return "laporan_versi" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo laporan (autoload).
func (d *VersiDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLaporan: {
			Domain:     "laporan",
			Type:       vernon.RelBelongsTo,
			FK:         VersiFieldLaporanID,
			LocalKey:   VersiFieldLaporanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"report_type", "period_label", "status"},
		},
	}
}

// Validate memvalidasi invariant laporan_versi sebelum write.
func (d *VersiDescriptor) Validate(data map[string]any) error {
	laporanID, _ := data[VersiFieldLaporanID].(string)
	if laporanID == "" {
		return errors.New("laporan_id wajib diisi")
	}

	versionNo, ok := data[VersiFieldVersionNo].(float64)
	if !ok || versionNo < 1 {
		return errors.New("version_number wajib diisi dan harus >= 1")
	}

	reportData := data[VersiFieldReportData]
	if reportData == nil {
		return errors.New("report_data wajib diisi")
	}

	reason, _ := data[VersiFieldCreatedReason].(string)
	if reason == "" {
		return errors.New("created_reason wajib diisi")
	}
	if !validCreatedReasons[reason] {
		return fmt.Errorf("created_reason tidak valid: %q", reason)
	}
	return nil
}
