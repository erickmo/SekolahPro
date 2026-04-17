// Package teaching_schedule adalah domain Vernon untuk jadwal mengajar.
//
// Teaching schedule terdiri dari dua tabel:
//   - time_slots: definisi slot waktu per hari (jam ke-1, istirahat, dll)
//   - schedule_entries: penempatan mata pelajaran ke slot waktu
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teaching_schedule

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field name constants ─────────────────────────────────────────────────────

// TimeSlot field names.
const (
	FieldName         = "name"
	FieldSlotNumber   = "slot_number"
	FieldSlotType     = "slot_type"
	FieldDayOfWeek    = "day_of_week"
	FieldSemesterTS   = "semester"
	FieldStartTime    = "start_time"
	FieldEndTime      = "end_time"
	FieldAcademicYearIDTS = "academic_year_id"
)

// ScheduleEntry field names.
const (
	FieldTimeSlotID     = "time_slot_id"
	FieldClassRoomIDSE  = "class_room_id"
	FieldSubjectIDSE    = "subject_id"
	FieldTeacherIDSE    = "teacher_id"
	FieldAcademicYearIDSE = "academic_year_id"
	FieldSemesterSE     = "semester"
)

// ── Enum constants ───────────────────────────────────────────────────────────

// Slot type constants.
const (
	SlotTypeLesson   = "lesson"
	SlotTypeBreak    = "break"
	SlotTypeAssembly = "assembly"
	SlotTypePrayer   = "prayer"
)

// Semester constants.
const (
	SemesterGanjil = "ganjil"
	SemesterGenap  = "genap"
)

// ── Range constants ──────────────────────────────────────────────────────────

const (
	SlotNumberMin = 1
	SlotNumberMax = 15
	DayOfWeekMin  = 1
	DayOfWeekMax  = 7
)

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelAcademicYearTS = "academic_year"
	RelTimeSlot       = "time_slot"
	RelClassRoomSE    = "class_room"
	RelSubjectSE      = "subject"
	RelTeacherSE      = "teacher"
	RelAcademicYearSE = "academic_year"
)

// ── Validation maps ──────────────────────────────────────────────────────────

var validSlotTypes = map[string]bool{
	SlotTypeLesson: true, SlotTypeBreak: true,
	SlotTypeAssembly: true, SlotTypePrayer: true,
}

var validSemesters = map[string]bool{
	SemesterGanjil: true, SemesterGenap: true,
}

// ── TimeSlotDescriptor ───────────────────────────────────────────────────────

// TimeSlotDescriptor mengimplementasi vernon.DomainDescriptor untuk time_slots.
type TimeSlotDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *TimeSlotDescriptor) TableName() string { return "time_slots" }

// DefaultRels mendefinisikan relasi domain time_slots.
// belongs_to academic_year dengan autoload.
func (d *TimeSlotDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelAcademicYearTS: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearIDTS,
			LocalKey:   FieldAcademicYearIDTS,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain time_slots sebelum write.
func (d *TimeSlotDescriptor) Validate(data map[string]any) error {
	if err := validateTimeSlotRequired(data); err != nil {
		return err
	}
	return validateTimeSlotEnums(data)
}

// validateTimeSlotRequired memeriksa field wajib time_slots.
func validateTimeSlotRequired(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	semester, _ := data[FieldSemesterTS].(string)
	if semester == "" {
		return errors.New("semester wajib diisi")
	}

	if _, ok := data[FieldStartTime]; !ok {
		return errors.New("start_time wajib diisi")
	}
	if _, ok := data[FieldEndTime]; !ok {
		return errors.New("end_time wajib diisi")
	}
	return nil
}

// validateTimeSlotEnums memeriksa enum dan range constraints time_slots.
func validateTimeSlotEnums(data map[string]any) error {
	slotType, _ := data[FieldSlotType].(string)
	if slotType != "" && !validSlotTypes[slotType] {
		return fmt.Errorf("slot_type tidak valid: %q (harus lesson/break/assembly/prayer)", slotType)
	}

	semester, _ := data[FieldSemesterTS].(string)
	if semester != "" && !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}

	if err := validateSlotNumberRange(data); err != nil {
		return err
	}

	if err := validateDayOfWeekRange(data); err != nil {
		return err
	}

	return validateTimeOrder(data)
}

// validateSlotNumberRange memeriksa slot_number 1-15.
func validateSlotNumberRange(data map[string]any) error {
	slotNum, ok := data[FieldSlotNumber].(float64)
	if !ok {
		return nil
	}
	if slotNum < SlotNumberMin || slotNum > SlotNumberMax {
		return fmt.Errorf("slot_number harus antara %d dan %d", SlotNumberMin, SlotNumberMax)
	}
	return nil
}

// validateDayOfWeekRange memeriksa day_of_week 1-7.
func validateDayOfWeekRange(data map[string]any) error {
	day, ok := data[FieldDayOfWeek].(float64)
	if !ok {
		return nil
	}
	if day < DayOfWeekMin || day > DayOfWeekMax {
		return fmt.Errorf("day_of_week harus antara %d dan %d", DayOfWeekMin, DayOfWeekMax)
	}
	return nil
}

// validateTimeOrder memeriksa start_time < end_time (format "HH:MM").
func validateTimeOrder(data map[string]any) error {
	startStr, _ := data[FieldStartTime].(string)
	endStr, _ := data[FieldEndTime].(string)
	if startStr == "" || endStr == "" {
		return nil
	}
	if startStr >= endStr {
		return fmt.Errorf("start_time (%s) harus lebih awal dari end_time (%s)", startStr, endStr)
	}
	return nil
}

// ── ScheduleEntryDescriptor ──────────────────────────────────────────────────

// ScheduleEntryDescriptor mengimplementasi vernon.DomainDescriptor untuk schedule_entries.
type ScheduleEntryDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ScheduleEntryDescriptor) TableName() string { return "schedule_entries" }

// DefaultRels mendefinisikan relasi domain schedule_entries.
// 5 BelongsTo autoload: time_slot, class_room, subject, teacher, academic_year.
func (d *ScheduleEntryDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTimeSlot: {
			Domain:     "time_slots",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTimeSlotID,
			LocalKey:   FieldTimeSlotID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "slot_number", "start_time", "end_time"},
		},
		RelClassRoomSE: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomIDSE,
			LocalKey:   FieldClassRoomIDSE,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
		RelSubjectSE: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSubjectIDSE,
			LocalKey:   FieldSubjectIDSE,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RelTeacherSE: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherIDSE,
			LocalKey:   FieldTeacherIDSE,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		RelAcademicYearSE: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearIDSE,
			LocalKey:   FieldAcademicYearIDSE,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain schedule_entries sebelum write.
func (d *ScheduleEntryDescriptor) Validate(data map[string]any) error {
	return validateScheduleEntryRequired(data)
}

// validateScheduleEntryRequired memeriksa semua field wajib schedule_entries.
func validateScheduleEntryRequired(data map[string]any) error {
	timeSlotID, _ := data[FieldTimeSlotID].(string)
	if timeSlotID == "" {
		return errors.New("time_slot_id wajib diisi")
	}

	classRoomID, _ := data[FieldClassRoomIDSE].(string)
	if classRoomID == "" {
		return errors.New("class_room_id wajib diisi")
	}

	subjectID, _ := data[FieldSubjectIDSE].(string)
	if subjectID == "" {
		return errors.New("subject_id wajib diisi")
	}

	teacherID, _ := data[FieldTeacherIDSE].(string)
	if teacherID == "" {
		return errors.New("teacher_id wajib diisi")
	}

	academicYearID, _ := data[FieldAcademicYearIDSE].(string)
	if academicYearID == "" {
		return errors.New("academic_year_id wajib diisi")
	}

	semester, _ := data[FieldSemesterSE].(string)
	if semester == "" {
		return errors.New("semester wajib diisi")
	}
	if !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}
	return nil
}
