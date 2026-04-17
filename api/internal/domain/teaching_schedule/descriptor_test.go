package teaching_schedule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teaching_schedule"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── TimeSlotDescriptor metadata tests ────────────────────────────────────────

func TestTimeSlotDescriptor_TableName(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	assert.Equal(t, "time_slots", d.TableName())
}

func TestTimeSlotDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teaching_schedule.TimeSlotDescriptor{}
}

func TestTimeSlotDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "time_slots should have 1 BelongsTo relation")
}

func TestTimeSlotDescriptor_DefaultRels_AcademicYear(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teaching_schedule.RelAcademicYearTS]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "academic_year should be autoloaded")
}

// ── TimeSlotDescriptor validation tests ──────────────────────────────────────

func validTimeSlotData() map[string]any {
	return map[string]any{
		teaching_schedule.FieldName:      "Jam Ke-1",
		teaching_schedule.FieldSlotNumber:    float64(1),
		teaching_schedule.FieldSlotType:      teaching_schedule.SlotTypeLesson,
		teaching_schedule.FieldDayOfWeek:     float64(1),
		teaching_schedule.FieldSemesterTS:    teaching_schedule.SemesterGanjil,
		teaching_schedule.FieldStartTime:     "07:30",
		teaching_schedule.FieldEndTime:       "08:15",
		teaching_schedule.FieldAcademicYearIDTS: "00000000-0000-0000-0000-000000000001",
	}
}

func TestTimeSlotDescriptor_Validate_Success(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	err := d.Validate(validTimeSlotData())
	assert.NoError(t, err)
}

func TestTimeSlotDescriptor_Validate_AllSlotTypes(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	types := []string{
		teaching_schedule.SlotTypeLesson, teaching_schedule.SlotTypeBreak,
		teaching_schedule.SlotTypeAssembly, teaching_schedule.SlotTypePrayer,
	}
	for _, st := range types {
		data := validTimeSlotData()
		data[teaching_schedule.FieldSlotType] = st
		err := d.Validate(data)
		assert.NoError(t, err, "slot_type=%q should be valid", st)
	}
}

func TestTimeSlotDescriptor_Validate_MissingName(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	delete(data, teaching_schedule.FieldName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "name wajib diisi")
}

func TestTimeSlotDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	delete(data, teaching_schedule.FieldSemesterTS)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestTimeSlotDescriptor_Validate_MissingStartTime(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	delete(data, teaching_schedule.FieldStartTime)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "start_time wajib diisi")
}

func TestTimeSlotDescriptor_Validate_MissingEndTime(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	delete(data, teaching_schedule.FieldEndTime)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "end_time wajib diisi")
}

func TestTimeSlotDescriptor_Validate_InvalidSlotType(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldSlotType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_type tidak valid")
}

func TestTimeSlotDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldSemesterTS] = "summer"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestTimeSlotDescriptor_Validate_SlotNumberOutOfRange(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldSlotNumber] = float64(20)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_number harus antara")
}

func TestTimeSlotDescriptor_Validate_SlotNumberZero(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldSlotNumber] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_number harus antara")
}

func TestTimeSlotDescriptor_Validate_DayOfWeekOutOfRange(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldDayOfWeek] = float64(8)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "day_of_week harus antara")
}

func TestTimeSlotDescriptor_Validate_StartTimeAfterEndTime(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldStartTime] = "10:00"
	data[teaching_schedule.FieldEndTime] = "09:00"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "start_time")
}

func TestTimeSlotDescriptor_Validate_StartTimeEqualsEndTime(t *testing.T) {
	d := &teaching_schedule.TimeSlotDescriptor{}
	data := validTimeSlotData()
	data[teaching_schedule.FieldStartTime] = "07:30"
	data[teaching_schedule.FieldEndTime] = "07:30"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "start_time")
}

// ── ScheduleEntryDescriptor metadata tests ───────────────────────────────────

func TestScheduleEntryDescriptor_TableName(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	assert.Equal(t, "schedule_entries", d.TableName())
}

func TestScheduleEntryDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teaching_schedule.ScheduleEntryDescriptor{}
}

func TestScheduleEntryDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 5, "schedule_entries should have 5 BelongsTo relations")
}

func TestScheduleEntryDescriptor_DefaultRels_AllAutoload(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	rels := d.DefaultRels()

	expected := []string{
		teaching_schedule.RelTimeSlot, teaching_schedule.RelClassRoomSE,
		teaching_schedule.RelSubjectSE, teaching_schedule.RelTeacherSE,
		teaching_schedule.RelAcademicYearSE,
	}
	for _, name := range expected {
		rel, ok := rels[name]
		require.True(t, ok, "should have %s relation", name)
		assert.Equal(t, vernon.RelBelongsTo, rel.Type, "%s should be belongs_to", name)
		assert.True(t, rel.IsAutoload, "%s should be autoloaded", name)
	}
}

// ── ScheduleEntryDescriptor validation tests ─────────────────────────────────

func validScheduleEntryData() map[string]any {
	return map[string]any{
		teaching_schedule.FieldTimeSlotID:       "00000000-0000-0000-0000-000000000001",
		teaching_schedule.FieldClassRoomIDSE:    "00000000-0000-0000-0000-000000000002",
		teaching_schedule.FieldSubjectIDSE:      "00000000-0000-0000-0000-000000000003",
		teaching_schedule.FieldTeacherIDSE:      "00000000-0000-0000-0000-000000000004",
		teaching_schedule.FieldAcademicYearIDSE: "00000000-0000-0000-0000-000000000005",
		teaching_schedule.FieldSemesterSE:       teaching_schedule.SemesterGanjil,
	}
}

func TestScheduleEntryDescriptor_Validate_Success(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	err := d.Validate(validScheduleEntryData())
	assert.NoError(t, err)
}

func TestScheduleEntryDescriptor_Validate_MissingTimeSlotID(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldTimeSlotID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "time_slot_id wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_MissingClassRoomID(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldClassRoomIDSE)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "class_room_id wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_MissingSubjectID(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldSubjectIDSE)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "subject_id wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldTeacherIDSE)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldAcademicYearIDSE)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	delete(data, teaching_schedule.FieldSemesterSE)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestScheduleEntryDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &teaching_schedule.ScheduleEntryDescriptor{}
	data := validScheduleEntryData()
	data[teaching_schedule.FieldSemesterSE] = "summer"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}
