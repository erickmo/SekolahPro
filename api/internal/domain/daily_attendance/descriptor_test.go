package daily_attendance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/daily_attendance"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// -- Descriptor metadata tests ---------------------------------------------------

func TestDescriptor_TableName(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	assert.Equal(t, "daily_attendances", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &daily_attendance.Descriptor{}
}

// -- DefaultRels tests -----------------------------------------------------------

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "daily_attendance should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[daily_attendance.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, daily_attendance.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestDescriptor_DefaultRels_ClassRoomRelation(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[daily_attendance.RelClassRoom]
	require.True(t, ok, "should have class_room relation")

	assert.Equal(t, "class_rooms", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, daily_attendance.FieldClassRoomID, rel.FK)
	assert.True(t, rel.IsAutoload, "class_room should be autoloaded")
	assert.Equal(t, []string{"name", "grade_level"}, rel.Fields)
}

// -- Validation tests ------------------------------------------------------------

func validDailyAttendanceData() map[string]any {
	return map[string]any{
		daily_attendance.FieldStudentID:      "00000000-0000-0000-0000-000000000001",
		daily_attendance.FieldClassRoomID:    "00000000-0000-0000-0000-000000000002",
		daily_attendance.FieldAttendanceDate: "2026-04-17",
		daily_attendance.FieldStatus:         daily_attendance.StatusPresent,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	err := d.Validate(validDailyAttendanceData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	statuses := []string{
		daily_attendance.StatusPresent, daily_attendance.StatusAbsent,
		daily_attendance.StatusLate, daily_attendance.StatusExcused,
		daily_attendance.StatusSick, daily_attendance.StatusPermit,
	}
	for _, status := range statuses {
		data := validDailyAttendanceData()
		data[daily_attendance.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	data := validDailyAttendanceData()
	delete(data, daily_attendance.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_MissingClassRoomID(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	data := validDailyAttendanceData()
	delete(data, daily_attendance.FieldClassRoomID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "class_room_id wajib diisi")
}

func TestDescriptor_Validate_MissingAttendanceDate(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	data := validDailyAttendanceData()
	delete(data, daily_attendance.FieldAttendanceDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "attendance_date wajib diisi")
}

func TestDescriptor_Validate_MissingStatus(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	data := validDailyAttendanceData()
	delete(data, daily_attendance.FieldStatus)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status wajib diisi")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &daily_attendance.Descriptor{}
	data := validDailyAttendanceData()
	data[daily_attendance.FieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}
