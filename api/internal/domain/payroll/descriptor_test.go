package payroll_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/payroll"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── ConfigDescriptor ──────────────────────────────────────────────────────────

func TestConfigDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "payroll_configs", (&payroll.ConfigDescriptor{}).TableName())
}

func TestConfigDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &payroll.ConfigDescriptor{}
}

func TestConfigDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&payroll.ConfigDescriptor{}).DefaultRels(), 1)
}

func TestConfigDescriptor_DefaultRels_Teacher(t *testing.T) {
	rel, ok := (&payroll.ConfigDescriptor{}).DefaultRels()[payroll.RelTeacher]
	require.True(t, ok)
	assert.Equal(t, "teachers", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

func validConfigData() map[string]any {
	return map[string]any{
		payroll.CfgFieldTeacherID:    "00000000-0000-0000-0000-000000000001",
		payroll.CfgFieldEmployeeType: payroll.EmpTypePNS,
		payroll.CfgFieldBaseSalary:   float64(5000000),
	}
}

func TestConfigDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&payroll.ConfigDescriptor{}).Validate(validConfigData()))
}

func TestConfigDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := validConfigData()
	delete(d, payroll.CfgFieldTeacherID)
	assert.ErrorContains(t, (&payroll.ConfigDescriptor{}).Validate(d), "teacher_id wajib diisi")
}

func TestConfigDescriptor_Validate_InvalidEmployeeType(t *testing.T) {
	d := validConfigData()
	d[payroll.CfgFieldEmployeeType] = "kontrak"
	assert.ErrorContains(t, (&payroll.ConfigDescriptor{}).Validate(d), "employee_type tidak valid")
}

func TestConfigDescriptor_Validate_MissingBaseSalary(t *testing.T) {
	d := validConfigData()
	delete(d, payroll.CfgFieldBaseSalary)
	assert.ErrorContains(t, (&payroll.ConfigDescriptor{}).Validate(d), "base_salary wajib diisi")
}

// ── PeriodDescriptor ──────────────────────────────────────────────────────────

func TestPeriodDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "payroll_periods", (&payroll.PeriodDescriptor{}).TableName())
}

func TestPeriodDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &payroll.PeriodDescriptor{}
}

func TestPeriodDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&payroll.PeriodDescriptor{}).DefaultRels(), 1)
}

func TestPeriodDescriptor_DefaultRels_AcademicYear(t *testing.T) {
	rel, ok := (&payroll.PeriodDescriptor{}).DefaultRels()[payroll.RelAcademicYear]
	require.True(t, ok)
	assert.Equal(t, "academic_years", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

func validPeriodData() map[string]any {
	return map[string]any{
		payroll.PerFieldName:           "April 2026",
		payroll.PerFieldAcademicYearID: "00000000-0000-0000-0000-000000000001",
		payroll.PerFieldMonth:          float64(4),
		payroll.PerFieldYear:           float64(2026),
		payroll.PerFieldStartDate:      "2026-04-01",
		payroll.PerFieldEndDate:        "2026-04-30",
	}
}

func TestPeriodDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&payroll.PeriodDescriptor{}).Validate(validPeriodData()))
}

func TestPeriodDescriptor_Validate_MissingPeriodName(t *testing.T) {
	d := validPeriodData()
	delete(d, payroll.PerFieldName)
	assert.ErrorContains(t, (&payroll.PeriodDescriptor{}).Validate(d), "period_name wajib diisi")
}

func TestPeriodDescriptor_Validate_MonthOutOfRange(t *testing.T) {
	d := validPeriodData()
	d[payroll.PerFieldMonth] = float64(13)
	assert.ErrorContains(t, (&payroll.PeriodDescriptor{}).Validate(d), "month harus antara 1 dan 12")
}

// ── EntryDescriptor ───────────────────────────────────────────────────────────

func TestEntryDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "payroll_entries", (&payroll.EntryDescriptor{}).TableName())
}

func TestEntryDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &payroll.EntryDescriptor{}
}

func TestEntryDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&payroll.EntryDescriptor{}).DefaultRels(), 2)
}

func validEntryData() map[string]any {
	return map[string]any{
		payroll.EntFieldPeriodID:      "00000000-0000-0000-0000-000000000001",
		payroll.EntFieldTeacherID:     "00000000-0000-0000-0000-000000000002",
		payroll.EntFieldPayslipNumber: "SLP-2026-04-001",
	}
}

func TestEntryDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&payroll.EntryDescriptor{}).Validate(validEntryData()))
}

func TestEntryDescriptor_Validate_MissingPeriodID(t *testing.T) {
	d := validEntryData()
	delete(d, payroll.EntFieldPeriodID)
	assert.ErrorContains(t, (&payroll.EntryDescriptor{}).Validate(d), "period_id wajib diisi")
}

func TestEntryDescriptor_Validate_MissingPayslipNumber(t *testing.T) {
	d := validEntryData()
	delete(d, payroll.EntFieldPayslipNumber)
	assert.ErrorContains(t, (&payroll.EntryDescriptor{}).Validate(d), "payslip_number wajib diisi")
}

// ── ComponentDescriptor ───────────────────────────────────────────────────────

func TestComponentDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "payroll_components", (&payroll.ComponentDescriptor{}).TableName())
}

func TestComponentDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &payroll.ComponentDescriptor{}
}

func TestComponentDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&payroll.ComponentDescriptor{}).DefaultRels(), 1)
}

func TestComponentDescriptor_DefaultRels_Entry(t *testing.T) {
	rel, ok := (&payroll.ComponentDescriptor{}).DefaultRels()[payroll.RelEntry]
	require.True(t, ok)
	assert.Equal(t, "payroll_entries", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

func validComponentData() map[string]any {
	return map[string]any{
		payroll.CompFieldEntryID:       "00000000-0000-0000-0000-000000000001",
		payroll.CompFieldComponentType: payroll.CompTypeEarning,
		payroll.CompFieldName:          "Tunjangan Transport",
	}
}

func TestComponentDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&payroll.ComponentDescriptor{}).Validate(validComponentData()))
}

func TestComponentDescriptor_Validate_MissingEntryID(t *testing.T) {
	d := validComponentData()
	delete(d, payroll.CompFieldEntryID)
	assert.ErrorContains(t, (&payroll.ComponentDescriptor{}).Validate(d), "entry_id wajib diisi")
}

func TestComponentDescriptor_Validate_InvalidComponentType(t *testing.T) {
	d := validComponentData()
	d[payroll.CompFieldComponentType] = "bonus"
	assert.ErrorContains(t, (&payroll.ComponentDescriptor{}).Validate(d), "component_type tidak valid")
}

func TestComponentDescriptor_Validate_MissingName(t *testing.T) {
	d := validComponentData()
	delete(d, payroll.CompFieldName)
	assert.ErrorContains(t, (&payroll.ComponentDescriptor{}).Validate(d), "name wajib diisi")
}
