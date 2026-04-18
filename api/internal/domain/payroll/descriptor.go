// Package payroll adalah domain Vernon untuk penggajian staff.
//
// Terdiri dari 4 tabel: payroll_configs, payroll_periods, payroll_entries, payroll_components.
// payroll_configs memiliki 1 BelongsTo autoload: teacher.
// payroll_periods memiliki 1 BelongsTo autoload: academic_year.
// payroll_entries memiliki 2 BelongsTo autoload: period, teacher.
// payroll_components memiliki 1 BelongsTo autoload: entry.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package payroll

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Config field constants ────────────────────────────────────────────────────

const (
	CfgFieldTeacherID         = "teacher_id"
	CfgFieldEmployeeType      = "employee_type"
	CfgFieldBaseSalary        = "base_salary"
	CfgFieldTransportAllow    = "transport_allowance"
	CfgFieldMealAllow         = "meal_allowance"
	CfgFieldPositionAllow     = "position_allowance"
	CfgFieldFamilyAllow       = "family_allowance"
	CfgFieldRiceAllow         = "rice_allowance"
	CfgFieldPphStatus         = "pph_status"
	CfgFieldBankName          = "bank_name"
	CfgFieldBankAccount       = "bank_account"
	CfgFieldIsActive          = "is_active"
)

// ── Period field constants ─────────────────────────────────────────────────────

const (
	PerFieldName           = "period_name"
	PerFieldAcademicYearID = "academic_year_id"
	PerFieldMonth          = "month"
	PerFieldYear           = "year"
	PerFieldStartDate      = "start_date"
	PerFieldEndDate        = "end_date"
	PerFieldWorkflowStatus = "workflow_status"
	PerFieldTotalEmployees = "total_employees"
	PerFieldTotalGross     = "total_gross"
	PerFieldTotalDeductions = "total_deductions"
	PerFieldTotalNet       = "total_net"
	PerFieldProcessedBy    = "processed_by"
	PerFieldProcessedAt    = "processed_at"
)

// ── Entry field constants ─────────────────────────────────────────────────────

const (
	EntFieldPeriodID        = "period_id"
	EntFieldTeacherID       = "teacher_id"
	EntFieldEmployeeType    = "employee_type"
	EntFieldBaseSalary      = "base_salary"
	EntFieldTotalAllowances = "total_allowances"
	EntFieldGrossSalary     = "gross_salary"
	EntFieldTotalDeductions = "total_deductions"
	EntFieldNetSalary       = "net_salary"
	EntFieldPph21Amount     = "pph21_amount"
	EntFieldWorkingDays     = "working_days"
	EntFieldPresentDays     = "present_days"
	EntFieldAbsentDays      = "absent_days"
	EntFieldLeaveDays       = "leave_days"
	EntFieldPayslipNumber   = "payslip_number"
	EntFieldNotes           = "notes"
)

// ── Component field constants ─────────────────────────────────────────────────

const (
	CompFieldEntryID            = "entry_id"
	CompFieldComponentType      = "component_type"
	CompFieldCategory           = "category"
	CompFieldName               = "name"
	CompFieldAmount             = "amount"
	CompFieldCalculationMethod  = "calculation_method"
	CompFieldIsRecurring        = "is_recurring"
	CompFieldReferenceID        = "reference_id"
	CompFieldNotes              = "notes"
)

// ── Enum constants ────────────────────────────────────────────────────────────

const (
	EmpTypePNS     = "pns"
	EmpTypeHonorer = "honorer"
	EmpTypeYayasan = "yayasan"
)

const (
	PphNonPKP = "non_pkp"
	PphPTKP   = "ptkp"
	PphPKP    = "pkp"
)

const (
	PWSayDraft       = "draft"
	PWSCalculating   = "calculating"
	PWSCalculated    = "calculated"
	PWSApproved      = "approved"
	PWSProcessing    = "processing"
	PWSPaid          = "paid"
	PWSCancelled     = "cancelled"
	PWSFailed        = "failed"
)

const (
	CompTypeEarning   = "earning"
	CompTypeDeduction = "deduction"
)

const (
	CalcFixed           = "fixed"
	CalcPercentage      = "percentage"
	CalcFormula         = "formula"
	CalcAttendanceBased = "attendance_based"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelTeacher       = "teacher"
	RelAcademicYear  = "academic_year"
	RelPeriod        = "period"
	RelEntry         = "entry"
)

var validEmployeeTypes = map[string]bool{
	EmpTypePNS: true, EmpTypeHonorer: true, EmpTypeYayasan: true,
}

var validPphStatuses = map[string]bool{
	PphNonPKP: true, PphPTKP: true, PphPKP: true,
}

var validPeriodWorkflowStatuses = map[string]bool{
	PWSayDraft: true, PWSCalculating: true, PWSCalculated: true,
	PWSApproved: true, PWSProcessing: true, PWSPaid: true,
	PWSCancelled: true, PWSFailed: true,
}

var validComponentTypes = map[string]bool{
	CompTypeEarning: true, CompTypeDeduction: true,
}

var validCalcMethods = map[string]bool{
	CalcFixed: true, CalcPercentage: true, CalcFormula: true, CalcAttendanceBased: true,
}

// ── ConfigDescriptor ──────────────────────────────────────────────────────────

// ConfigDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_configs.
type ConfigDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ConfigDescriptor) TableName() string { return "payroll_configs" }

// DefaultRels mendefinisikan relasi payroll_configs: teacher.
func (d *ConfigDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         CfgFieldTeacherID,
			LocalKey:   CfgFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type"},
		},
	}
}

// Validate memvalidasi invariant payroll_configs.
func (d *ConfigDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, CfgFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := requireEnum(data, CfgFieldEmployeeType, "employee_type", validEmployeeTypes); err != nil {
		return err
	}
	salary, _ := data[CfgFieldBaseSalary]
	if salary == nil {
		return fmt.Errorf("base_salary wajib diisi")
	}
	return nil
}

// ── PeriodDescriptor ──────────────────────────────────────────────────────────

// PeriodDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_periods.
type PeriodDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PeriodDescriptor) TableName() string { return "payroll_periods" }

// DefaultRels mendefinisikan relasi payroll_periods: academic_year.
func (d *PeriodDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         PerFieldAcademicYearID,
			LocalKey:   PerFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant payroll_periods.
func (d *PeriodDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, PerFieldName, "period_name"); err != nil {
		return err
	}
	if err := requireString(data, PerFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	if err := requireIntRange(data, PerFieldMonth, "month", 1, 12); err != nil {
		return err
	}
	if err := requireString(data, PerFieldStartDate, "start_date"); err != nil {
		return err
	}
	return requireString(data, PerFieldEndDate, "end_date")
}

// ── EntryDescriptor ───────────────────────────────────────────────────────────

// EntryDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_entries.
type EntryDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *EntryDescriptor) TableName() string { return "payroll_entries" }

// DefaultRels mendefinisikan relasi payroll_entries: period + teacher.
func (d *EntryDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelPeriod: {
			Domain:     "payroll_periods",
			Type:       vernon.RelBelongsTo,
			FK:         EntFieldPeriodID,
			LocalKey:   EntFieldPeriodID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"period_name", "month", "year", "workflow_status"},
		},
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         EntFieldTeacherID,
			LocalKey:   EntFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
	}
}

// Validate memvalidasi invariant payroll_entries.
func (d *EntryDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, EntFieldPeriodID, "period_id"); err != nil {
		return err
	}
	if err := requireString(data, EntFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	return requireString(data, EntFieldPayslipNumber, "payslip_number")
}

// ── ComponentDescriptor ───────────────────────────────────────────────────────

// ComponentDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_components.
type ComponentDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ComponentDescriptor) TableName() string { return "payroll_components" }

// DefaultRels mendefinisikan relasi payroll_components: entry.
func (d *ComponentDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelEntry: {
			Domain:     "payroll_entries",
			Type:       vernon.RelBelongsTo,
			FK:         CompFieldEntryID,
			LocalKey:   CompFieldEntryID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"payslip_number", "teacher_id"},
		},
	}
}

// Validate memvalidasi invariant payroll_components.
func (d *ComponentDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, CompFieldEntryID, "entry_id"); err != nil {
		return err
	}
	if err := requireEnum(data, CompFieldComponentType, "component_type", validComponentTypes); err != nil {
		return err
	}
	return requireString(data, CompFieldName, "name")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func requireEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}

func requireIntRange(data map[string]any, field, label string, min, max int) error {
	val, ok := data[field]
	if !ok {
		return fmt.Errorf("%s wajib diisi", label)
	}
	var intVal int
	switch v := val.(type) {
	case int:
		intVal = v
	case float64:
		intVal = int(v)
	default:
		return fmt.Errorf("%s harus berupa angka", label)
	}
	if intVal < min || intVal > max {
		return fmt.Errorf("%s harus antara %d dan %d", label, min, max)
	}
	return nil
}
