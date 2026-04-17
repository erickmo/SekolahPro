// Package leave_management adalah domain Vernon untuk manajemen cuti dan izin guru/staff.
//
// Terdiri dari 4 deskriptor:
//   - LeaveTypeDescriptor     : konfigurasi jenis cuti (standalone, tanpa BelongsTo)
//   - LeaveBalanceDescriptor  : saldo cuti per guru per tahun (3 BelongsTo)
//   - LeaveRequestDescriptor  : pengajuan cuti (4 BelongsTo)
//   - LeaveApprovalLogDescriptor : riwayat approval (1 BelongsTo)
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package leave_management

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Leave Type ──────────────────────────────────────────────────────────────

// Field name constants — leave_types.
const (
	LTFieldCode                = "code"
	LTFieldName                = "name"
	LTFieldDescription         = "description"
	LTFieldMaxDaysPerYear      = "max_days_per_year"
	LTFieldIsPaid              = "is_paid"
	LTFieldApplicableTo        = "applicable_to"
	LTFieldApprovalLevels      = "approval_levels"
	LTFieldRequiresDocument    = "requires_document"
	LTFieldIsActive            = "is_active"
)

// Leave type code enum constants.
const (
	CodeCutiTahunan    = "cuti_tahunan"
	CodeCutiSakit      = "cuti_sakit"
	CodeCutiMelahirkan = "cuti_melahirkan"
	CodeCutiBesar      = "cuti_besar"
	CodeIzin           = "izin"
	CodeTugasBelajar   = "tugas_belajar"
)

// ApplicableTo enum constants.
const (
	ApplicableAll     = "all"
	ApplicablePNS     = "pns"
	ApplicableHonorer = "honorer"
	ApplicableYayasan = "yayasan"
)

var validLeaveTypeCodes = map[string]bool{
	CodeCutiTahunan: true, CodeCutiSakit: true, CodeCutiMelahirkan: true,
	CodeCutiBesar: true, CodeIzin: true, CodeTugasBelajar: true,
}

var validApplicableTo = map[string]bool{
	ApplicableAll: true, ApplicablePNS: true, ApplicableHonorer: true, ApplicableYayasan: true,
}

// LeaveTypeDescriptor mengimplementasi vernon.DomainDescriptor untuk leave_types.
type LeaveTypeDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LeaveTypeDescriptor) TableName() string { return "leave_types" }

// DefaultRels — leave_types tidak memiliki relasi (standalone config).
func (d *LeaveTypeDescriptor) DefaultRels() map[string]vernon.RelDef { return nil }

// Validate memvalidasi invariant leave_types sebelum write.
func (d *LeaveTypeDescriptor) Validate(data map[string]any) error {
	code, _ := data[LTFieldCode].(string)
	if code == "" {
		return fmt.Errorf("code wajib diisi")
	}
	if !validLeaveTypeCodes[code] {
		return fmt.Errorf("code tidak valid: %q", code)
	}

	name, _ := data[LTFieldName].(string)
	if name == "" {
		return fmt.Errorf("name wajib diisi")
	}

	applicableTo, _ := data[LTFieldApplicableTo].(string)
	if applicableTo != "" && !validApplicableTo[applicableTo] {
		return fmt.Errorf("applicable_to tidak valid: %q", applicableTo)
	}

	return validateApprovalLevels(data)
}

func validateApprovalLevels(data map[string]any) error {
	levels, ok := data[LTFieldApprovalLevels]
	if !ok || levels == nil {
		return nil
	}
	levelFloat, ok := levels.(float64)
	if !ok {
		return nil
	}
	if levelFloat < 1 || levelFloat > 4 {
		return fmt.Errorf("approval_levels harus antara 1 dan 4")
	}
	return nil
}

// ── Leave Balance ───────────────────────────────────────────────────────────

// Field name constants — leave_balances.
const (
	LBFieldTeacherID      = "teacher_id"
	LBFieldLeaveTypeID    = "leave_type_id"
	LBFieldAcademicYearID = "academic_year_id"
	LBFieldYear           = "year"
	LBFieldInitialBalance = "initial_balance"
	LBFieldUsed           = "used"
	LBFieldRemaining      = "remaining"
	LBFieldCarryOver      = "carry_over"
)

// Relation name constants — leave_balances.
const (
	RelTeacher      = "teacher"
	RelLeaveType    = "leave_type"
	RelAcademicYear = "academic_year"
)

// LeaveBalanceDescriptor mengimplementasi vernon.DomainDescriptor untuk leave_balances.
type LeaveBalanceDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LeaveBalanceDescriptor) TableName() string { return "leave_balances" }

// DefaultRels mendefinisikan relasi domain ini.
// 3 BelongsTo autoload: teacher, leave_type, academic_year.
func (d *LeaveBalanceDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         LBFieldTeacherID,
			LocalKey:   LBFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type"},
		},
		RelLeaveType: {
			Domain:     "leave_types",
			Type:       vernon.RelBelongsTo,
			FK:         LBFieldLeaveTypeID,
			LocalKey:   LBFieldLeaveTypeID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"code", "name", "max_days_per_year"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         LBFieldAcademicYearID,
			LocalKey:   LBFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant leave_balances sebelum write.
func (d *LeaveBalanceDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, LBFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LBFieldLeaveTypeID, "leave_type_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LBFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}

	yearVal, ok := data[LBFieldYear]
	if !ok || yearVal == nil {
		return fmt.Errorf("year wajib diisi")
	}
	yearFloat, ok := yearVal.(float64)
	if !ok || yearFloat < 2020 || yearFloat > 2100 {
		return fmt.Errorf("year harus antara 2020 dan 2100")
	}

	return validateRequiredNonNegative(data, LBFieldInitialBalance, "initial_balance")
}

// ── Leave Request ───────────────────────────────────────────────────────────

// Field name constants — leave_requests.
const (
	LRFieldTeacherID             = "teacher_id"
	LRFieldLeaveTypeID           = "leave_type_id"
	LRFieldLeaveBalanceID        = "leave_balance_id"
	LRFieldAcademicYearID        = "academic_year_id"
	LRFieldStartDate             = "start_date"
	LRFieldEndDate               = "end_date"
	LRFieldTotalDays             = "total_days"
	LRFieldReason                = "reason"
	LRFieldStatus                = "status"
	LRFieldDocumentURL           = "document_url"
	LRFieldApprovalNotes         = "approval_notes"
	LRFieldCurrentApprovalLevel  = "current_approval_level"
	LRFieldNotes                 = "notes"
)

// Status enum constants — leave_requests.
const (
	StatusDraft           = "draft"
	StatusSubmitted       = "submitted"
	StatusPendingApproval = "pending_approval"
	StatusApproved        = "approved"
	StatusRejected        = "rejected"
	StatusCancelled       = "cancelled"
	StatusCompleted       = "completed"
)

// Relation name constants — leave_requests.
const (
	RelTeacherLR      = "teacher"
	RelLeaveTypeLR    = "leave_type"
	RelLeaveBalanceLR = "leave_balance"
	RelAcademicYearLR = "academic_year"
)

var validLeaveStatuses = map[string]bool{
	StatusDraft: true, StatusSubmitted: true, StatusPendingApproval: true,
	StatusApproved: true, StatusRejected: true, StatusCancelled: true, StatusCompleted: true,
}

// LeaveRequestDescriptor mengimplementasi vernon.DomainDescriptor untuk leave_requests.
type LeaveRequestDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LeaveRequestDescriptor) TableName() string { return "leave_requests" }

// DefaultRels mendefinisikan relasi domain ini.
// 4 BelongsTo autoload: teacher, leave_type, leave_balance, academic_year.
func (d *LeaveRequestDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacherLR: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         LRFieldTeacherID,
			LocalKey:   LRFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type", "role"},
		},
		RelLeaveTypeLR: {
			Domain:     "leave_types",
			Type:       vernon.RelBelongsTo,
			FK:         LRFieldLeaveTypeID,
			LocalKey:   LRFieldLeaveTypeID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"code", "name", "max_days_per_year", "approval_chain"},
		},
		RelLeaveBalanceLR: {
			Domain:     "leave_balances",
			Type:       vernon.RelBelongsTo,
			FK:         LRFieldLeaveBalanceID,
			LocalKey:   LRFieldLeaveBalanceID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"initial_balance", "used", "remaining"},
		},
		RelAcademicYearLR: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         LRFieldAcademicYearID,
			LocalKey:   LRFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant leave_requests sebelum write.
func (d *LeaveRequestDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, LRFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LRFieldLeaveTypeID, "leave_type_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LRFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	return validateLeaveRequestFields(data)
}

func validateLeaveRequestFields(data map[string]any) error {
	if err := validateRequiredString(data, LRFieldStartDate, "start_date"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LRFieldEndDate, "end_date"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LRFieldReason, "reason"); err != nil {
		return err
	}

	totalDaysVal, ok := data[LRFieldTotalDays]
	if !ok || totalDaysVal == nil {
		return fmt.Errorf("total_days wajib diisi")
	}
	totalDaysFloat, ok := totalDaysVal.(float64)
	if !ok || totalDaysFloat <= 0 {
		return fmt.Errorf("total_days harus lebih besar dari 0")
	}

	status, _ := data[LRFieldStatus].(string)
	if status != "" && !validLeaveStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ── Leave Approval Log ──────────────────────────────────────────────────────

// Field name constants — leave_approval_logs.
const (
	LALFieldLeaveRequestID = "leave_request_id"
	LALFieldApproverID     = "approver_id"
	LALFieldApprovalLevel  = "approval_level"
	LALFieldApproverRole   = "approver_role"
	LALFieldAction         = "action"
	LALFieldNotes          = "notes"
)

// Action enum constants — leave_approval_logs.
const (
	ActionApproved = "approved"
	ActionRejected = "rejected"
	ActionReturned = "returned"
)

// Approver role enum constants.
const (
	RoleWakilKepsek      = "wakil_kepsek"
	RoleKepalaSekolah    = "kepala_sekolah"
	RoleYayasan          = "yayasan"
	RoleDinasPendidikan  = "dinas_pendidikan"
)

// Relation name constants — leave_approval_logs.
const (
	RelLeaveRequest = "leave_request"
	RelApprover     = "approver"
)

var validApprovalActions = map[string]bool{
	ActionApproved: true, ActionRejected: true, ActionReturned: true,
}

var validApproverRoles = map[string]bool{
	RoleWakilKepsek: true, RoleKepalaSekolah: true,
	RoleYayasan: true, RoleDinasPendidikan: true,
}

// LeaveApprovalLogDescriptor mengimplementasi vernon.DomainDescriptor untuk leave_approval_logs.
type LeaveApprovalLogDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LeaveApprovalLogDescriptor) TableName() string { return "leave_approval_logs" }

// DefaultRels mendefinisikan relasi domain ini.
// 1 BelongsTo autoload: leave_request.
func (d *LeaveApprovalLogDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLeaveRequest: {
			Domain:     "leave_requests",
			Type:       vernon.RelBelongsTo,
			FK:         LALFieldLeaveRequestID,
			LocalKey:   LALFieldLeaveRequestID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"teacher_name", "leave_type", "start_date", "end_date", "total_days"},
		},
		RelApprover: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         LALFieldApproverID,
			LocalKey:   LALFieldApproverID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "role"},
		},
	}
}

// Validate memvalidasi invariant leave_approval_logs sebelum write.
func (d *LeaveApprovalLogDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, LALFieldLeaveRequestID, "leave_request_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, LALFieldApproverID, "approver_id"); err != nil {
		return err
	}

	levelVal, ok := data[LALFieldApprovalLevel]
	if !ok || levelVal == nil {
		return fmt.Errorf("approval_level wajib diisi")
	}
	levelFloat, ok := levelVal.(float64)
	if !ok || levelFloat < 1 || levelFloat > 4 {
		return fmt.Errorf("approval_level harus antara 1 dan 4")
	}

	role, _ := data[LALFieldApproverRole].(string)
	if role == "" {
		return fmt.Errorf("approver_role wajib diisi")
	}
	if !validApproverRoles[role] {
		return fmt.Errorf("approver_role tidak valid: %q", role)
	}

	action, _ := data[LALFieldAction].(string)
	if action == "" {
		return fmt.Errorf("action wajib diisi")
	}
	if !validApprovalActions[action] {
		return fmt.Errorf("action tidak valid: %q", action)
	}
	return nil
}

// ── Shared helpers ──────────────────────────────────────────────────────────

func validateRequiredString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func validateRequiredNonNegative(data map[string]any, field, label string) error {
	val, ok := data[field]
	if !ok || val == nil {
		return fmt.Errorf("%s wajib diisi", label)
	}
	num, ok := val.(float64)
	if !ok || num < 0 {
		return fmt.Errorf("%s harus berupa angka tidak negatif", label)
	}
	return nil
}
