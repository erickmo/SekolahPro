package teacher_attendance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teacher_attendance"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ──────────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	assert.Equal(t, "teacher_attendances", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_attendance.Descriptor{}
}

// ── Descriptor DefaultRels tests ───────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "teacher_attendances should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_attendance.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_attendance.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip", "employee_type", "role"}, rel.Fields)
}

func TestDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_attendance.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_attendance.FieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload, "academic_year should be autoloaded")
	assert.Equal(t, []string{"name"}, rel.Fields)
}

// ── Descriptor Validation tests ────────────────────────────────────────────────

func validAttendanceData() map[string]any {
	return map[string]any{
		teacher_attendance.FieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		teacher_attendance.FieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		teacher_attendance.FieldAttendanceDate: "2026-04-17",
		teacher_attendance.FieldStatus:         teacher_attendance.AttStatusPresent,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	err := d.Validate(validAttendanceData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	statuses := []string{
		teacher_attendance.AttStatusPresent, teacher_attendance.AttStatusSick,
		teacher_attendance.AttStatusPermitted, teacher_attendance.AttStatusAbsent,
		teacher_attendance.AttStatusDinasLuar, teacher_attendance.AttStatusCuti,
		teacher_attendance.AttStatusLibur,
	}
	for _, status := range statuses {
		data := validAttendanceData()
		data[teacher_attendance.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	delete(data, teacher_attendance.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	delete(data, teacher_attendance.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestDescriptor_Validate_MissingAttendanceDate(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	delete(data, teacher_attendance.FieldAttendanceDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "attendance_date wajib diisi")
}

func TestDescriptor_Validate_MissingStatus(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	delete(data, teacher_attendance.FieldStatus)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status wajib diisi")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	data[teacher_attendance.FieldStatus] = "late"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_InvalidClockInMethod(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	data[teacher_attendance.FieldClockInMethod] = "sms"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "clock_in_method tidak valid")
}

func TestDescriptor_Validate_InvalidClockOutMethod(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	data[teacher_attendance.FieldClockOutMethod] = "phone"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "clock_out_method tidak valid")
}

func TestDescriptor_Validate_OptionalClockMethod(t *testing.T) {
	d := &teacher_attendance.Descriptor{}
	data := validAttendanceData()
	// No clock_in_method should be fine
	err := d.Validate(data)
	assert.NoError(t, err)
}

// ── ConfigDescriptor metadata tests ────────────────────────────────────────────

func TestConfigDescriptor_TableName(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	assert.Equal(t, "teacher_attendance_configs", d.TableName())
}

func TestConfigDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_attendance.ConfigDescriptor{}
}

func TestConfigDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "config table should have no relations")
}

// ── ConfigDescriptor Validation tests ──────────────────────────────────────────

func validConfigData() map[string]any {
	return map[string]any{
		teacher_attendance.FieldWorkStartTime:        "07:00",
		teacher_attendance.FieldWorkEndTime:          "14:00",
		teacher_attendance.FieldLateToleranceMinutes: float64(15),
	}
}

func TestConfigDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	err := d.Validate(validConfigData())
	assert.NoError(t, err)
}

func TestConfigDescriptor_Validate_MissingWorkStartTime(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	data := validConfigData()
	delete(data, teacher_attendance.FieldWorkStartTime)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "work_start_time wajib diisi")
}

func TestConfigDescriptor_Validate_MissingWorkEndTime(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	data := validConfigData()
	delete(data, teacher_attendance.FieldWorkEndTime)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "work_end_time wajib diisi")
}

func TestConfigDescriptor_Validate_MissingTolerance(t *testing.T) {
	d := &teacher_attendance.ConfigDescriptor{}
	data := validConfigData()
	delete(data, teacher_attendance.FieldLateToleranceMinutes)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "late_tolerance_minutes wajib diisi")
}
