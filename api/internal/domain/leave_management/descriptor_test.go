package leave_management_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/leave_management"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── LeaveType: metadata tests ───────────────────────────────────────────────

func TestLeaveTypeDescriptor_TableName(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	assert.Equal(t, "leave_types", d.TableName())
}

func TestLeaveTypeDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &leave_management.LeaveTypeDescriptor{}
}

func TestLeaveTypeDescriptor_DefaultRels_Nil(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	rels := d.DefaultRels()
	assert.Nil(t, rels, "leave_types tidak memiliki relasi")
}

// ── LeaveType: validation tests ─────────────────────────────────────────────

func validLeaveTypeData() map[string]any {
	return map[string]any{
		leave_management.LTFieldCode: leave_management.CodeCutiTahunan,
		leave_management.LTFieldName: "Cuti Tahunan",
	}
}

func TestLeaveTypeDescriptor_Validate_Success(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	err := d.Validate(validLeaveTypeData())
	assert.NoError(t, err)
}

func TestLeaveTypeDescriptor_Validate_AllCodes(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	codes := []string{
		leave_management.CodeCutiTahunan, leave_management.CodeCutiSakit,
		leave_management.CodeCutiMelahirkan, leave_management.CodeCutiBesar,
		leave_management.CodeIzin, leave_management.CodeTugasBelajar,
	}
	for _, code := range codes {
		data := validLeaveTypeData()
		data[leave_management.LTFieldCode] = code
		err := d.Validate(data)
		assert.NoError(t, err, "code=%q should be valid", code)
	}
}

func TestLeaveTypeDescriptor_Validate_MissingCode(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	data := validLeaveTypeData()
	delete(data, leave_management.LTFieldCode)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "code wajib diisi")
}

func TestLeaveTypeDescriptor_Validate_InvalidCode(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	data := validLeaveTypeData()
	data[leave_management.LTFieldCode] = "cuti_liburan"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "code tidak valid")
}

func TestLeaveTypeDescriptor_Validate_MissingName(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	data := validLeaveTypeData()
	delete(data, leave_management.LTFieldName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "name wajib diisi")
}

func TestLeaveTypeDescriptor_Validate_InvalidApplicableTo(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	data := validLeaveTypeData()
	data[leave_management.LTFieldApplicableTo] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "applicable_to tidak valid")
}

func TestLeaveTypeDescriptor_Validate_ApprovalLevelsOutOfRange(t *testing.T) {
	d := &leave_management.LeaveTypeDescriptor{}
	data := validLeaveTypeData()
	data[leave_management.LTFieldApprovalLevels] = float64(5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "approval_levels harus antara 1 dan 4")
}

// ── LeaveBalance: metadata tests ────────────────────────────────────────────

func TestLeaveBalanceDescriptor_TableName(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	assert.Equal(t, "leave_balances", d.TableName())
}

func TestLeaveBalanceDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &leave_management.LeaveBalanceDescriptor{}
}

func TestLeaveBalanceDescriptor_DefaultRels_Count(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "leave_balances should have 3 BelongsTo relations")
}

func TestLeaveBalanceDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[leave_management.RelTeacher]
	require.True(t, ok, "should have teacher relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, leave_management.LBFieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestLeaveBalanceDescriptor_DefaultRels_LeaveTypeRelation(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[leave_management.RelLeaveType]
	require.True(t, ok, "should have leave_type relation")
	assert.Equal(t, "leave_types", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

func TestLeaveBalanceDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[leave_management.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")
	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

// ── LeaveBalance: validation tests ──────────────────────────────────────────

func validLeaveBalanceData() map[string]any {
	return map[string]any{
		leave_management.LBFieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		leave_management.LBFieldLeaveTypeID:    "00000000-0000-0000-0000-000000000002",
		leave_management.LBFieldAcademicYearID: "00000000-0000-0000-0000-000000000003",
		leave_management.LBFieldYear:           float64(2026),
		leave_management.LBFieldInitialBalance: float64(12),
	}
}

func TestLeaveBalanceDescriptor_Validate_Success(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	err := d.Validate(validLeaveBalanceData())
	assert.NoError(t, err)
}

func TestLeaveBalanceDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	data := validLeaveBalanceData()
	delete(data, leave_management.LBFieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestLeaveBalanceDescriptor_Validate_YearOutOfRange(t *testing.T) {
	d := &leave_management.LeaveBalanceDescriptor{}
	data := validLeaveBalanceData()
	data[leave_management.LBFieldYear] = float64(2019)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "year harus antara 2020 dan 2100")
}

// ── LeaveRequest: metadata tests ────────────────────────────────────────────

func TestLeaveRequestDescriptor_TableName(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	assert.Equal(t, "leave_requests", d.TableName())
}

func TestLeaveRequestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &leave_management.LeaveRequestDescriptor{}
}

func TestLeaveRequestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "leave_requests should have 4 BelongsTo relations")
}

func TestLeaveRequestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[leave_management.RelTeacherLR]
	require.True(t, ok, "should have teacher relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, leave_management.LRFieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"full_name", "nip", "employee_type", "role"}, rel.Fields)
}

func TestLeaveRequestDescriptor_DefaultRels_LeaveBalanceRelation(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[leave_management.RelLeaveBalanceLR]
	require.True(t, ok, "should have leave_balance relation")
	assert.Equal(t, "leave_balances", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

// ── LeaveRequest: validation tests ──────────────────────────────────────────

func validLeaveRequestData() map[string]any {
	return map[string]any{
		leave_management.LRFieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		leave_management.LRFieldLeaveTypeID:    "00000000-0000-0000-0000-000000000002",
		leave_management.LRFieldAcademicYearID: "00000000-0000-0000-0000-000000000003",
		leave_management.LRFieldStartDate:      "2026-04-20",
		leave_management.LRFieldEndDate:        "2026-04-22",
		leave_management.LRFieldTotalDays:      float64(3),
		leave_management.LRFieldReason:         "Cuti tahunan",
		leave_management.LRFieldStatus:         leave_management.StatusDraft,
	}
}

func TestLeaveRequestDescriptor_Validate_Success(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	err := d.Validate(validLeaveRequestData())
	assert.NoError(t, err)
}

func TestLeaveRequestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	statuses := []string{
		leave_management.StatusDraft, leave_management.StatusSubmitted,
		leave_management.StatusPendingApproval, leave_management.StatusApproved,
		leave_management.StatusRejected, leave_management.StatusCancelled,
		leave_management.StatusCompleted,
	}
	for _, status := range statuses {
		data := validLeaveRequestData()
		data[leave_management.LRFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestLeaveRequestDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	data := validLeaveRequestData()
	delete(data, leave_management.LRFieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestLeaveRequestDescriptor_Validate_MissingReason(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	data := validLeaveRequestData()
	delete(data, leave_management.LRFieldReason)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "reason wajib diisi")
}

func TestLeaveRequestDescriptor_Validate_TotalDaysZero(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	data := validLeaveRequestData()
	data[leave_management.LRFieldTotalDays] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_days harus lebih besar dari 0")
}

func TestLeaveRequestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	data := validLeaveRequestData()
	data[leave_management.LRFieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

// ── LeaveApprovalLog: metadata tests ────────────────────────────────────────

func TestLeaveApprovalLogDescriptor_TableName(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	assert.Equal(t, "leave_approval_logs", d.TableName())
}

func TestLeaveApprovalLogDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &leave_management.LeaveApprovalLogDescriptor{}
}

func TestLeaveApprovalLogDescriptor_DefaultRels_Count(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "leave_approval_logs should have 2 BelongsTo relations")
}

// ── LeaveApprovalLog: validation tests ──────────────────────────────────────

func validApprovalLogData() map[string]any {
	return map[string]any{
		leave_management.LALFieldLeaveRequestID: "00000000-0000-0000-0000-000000000001",
		leave_management.LALFieldApproverID:     "00000000-0000-0000-0000-000000000002",
		leave_management.LALFieldApprovalLevel:  float64(1),
		leave_management.LALFieldApproverRole:   leave_management.RoleKepalaSekolah,
		leave_management.LALFieldAction:         leave_management.ActionApproved,
	}
}

func TestLeaveApprovalLogDescriptor_Validate_Success(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	err := d.Validate(validApprovalLogData())
	assert.NoError(t, err)
}

func TestLeaveApprovalLogDescriptor_Validate_AllActions(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	actions := []string{
		leave_management.ActionApproved, leave_management.ActionRejected,
		leave_management.ActionReturned,
	}
	for _, action := range actions {
		data := validApprovalLogData()
		data[leave_management.LALFieldAction] = action
		err := d.Validate(data)
		assert.NoError(t, err, "action=%q should be valid", action)
	}
}

func TestLeaveApprovalLogDescriptor_Validate_AllRoles(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	roles := []string{
		leave_management.RoleWakilKepsek, leave_management.RoleKepalaSekolah,
		leave_management.RoleYayasan, leave_management.RoleDinasPendidikan,
	}
	for _, role := range roles {
		data := validApprovalLogData()
		data[leave_management.LALFieldApproverRole] = role
		err := d.Validate(data)
		assert.NoError(t, err, "role=%q should be valid", role)
	}
}

func TestLeaveApprovalLogDescriptor_Validate_MissingRequestID(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	data := validApprovalLogData()
	delete(data, leave_management.LALFieldLeaveRequestID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "leave_request_id wajib diisi")
}

func TestLeaveApprovalLogDescriptor_Validate_InvalidAction(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	data := validApprovalLogData()
	data[leave_management.LALFieldAction] = "forwarded"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "action tidak valid")
}

func TestLeaveApprovalLogDescriptor_Validate_InvalidRole(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	data := validApprovalLogData()
	data[leave_management.LALFieldApproverRole] = "guru_mapel"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "approver_role tidak valid")
}

func TestLeaveApprovalLogDescriptor_Validate_LevelOutOfRange(t *testing.T) {
	d := &leave_management.LeaveApprovalLogDescriptor{}
	data := validApprovalLogData()
	data[leave_management.LALFieldApprovalLevel] = float64(5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "approval_level harus antara 1 dan 4")
}
