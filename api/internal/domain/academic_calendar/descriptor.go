// Package academic_calendar adalah domain Vernon untuk kalender akademik.
//
// Satu tabel: academic_calendar_events.
// Setiap event belongs_to academic_year dan memiliki tipe serta rentang tanggal.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package academic_calendar

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	ACFieldName         = "name"
	ACFieldEventType    = "event_type"
	ACFieldStartDate    = "start_date"
	ACFieldEndDate      = "end_date"
	ACFieldSemester     = "semester"
	ACFieldColorCode    = "color_code"
	ACFieldDescription  = "description"
	ACFieldLocation     = "location"
	ACFieldAcademicYearID = "academic_year_id"
)

// Event type constants.
const (
	EventTypeHoliday         = "holiday"
	EventTypeExamPeriod      = "exam_period"
	EventTypeSchoolEvent     = "school_event"
	EventTypeSemesterStart   = "semester_start"
	EventTypeSemesterEnd     = "semester_end"
	EventTypeGraduation      = "graduation"
	EventTypeEnrollmentPeriod = "enrollment_period"
	EventTypeTeacherTraining = "teacher_training"
	EventTypeIslamicHoliday  = "islamic_holiday"
	EventTypePesantrenEvent  = "pesantren_event"
	EventTypeRamadanSchedule = "ramadan_schedule"
)

// Semester constants.
const (
	ACSemesterGanjil = "ganjil"
	ACSemesterGenap  = "genap"
)

// Relation name constants.
const (
	ACRelAcademicYear = "academic_year"
)

const (
	// dateFormat digunakan untuk parsing start_date/end_date.
	dateFormat = "2006-01-02"
)

var validEventTypes = map[string]bool{
	EventTypeHoliday: true, EventTypeExamPeriod: true,
	EventTypeSchoolEvent: true, EventTypeSemesterStart: true,
	EventTypeSemesterEnd: true, EventTypeGraduation: true,
	EventTypeEnrollmentPeriod: true, EventTypeTeacherTraining: true,
	EventTypeIslamicHoliday: true, EventTypePesantrenEvent: true,
	EventTypeRamadanSchedule: true,
}

var validACSemesters = map[string]bool{
	ACSemesterGanjil: true, ACSemesterGenap: true,
}

// colorCodeRegex memvalidasi format warna hex #RRGGBB.
var colorCodeRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// AcademicCalendarEventsDescriptor mengimplementasi vernon.DomainDescriptor untuk academic_calendar_events.
type AcademicCalendarEventsDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *AcademicCalendarEventsDescriptor) TableName() string { return "academic_calendar_events" }

// DefaultRels mendefinisikan relasi domain academic_calendar_events.
// belongs_to academic_year (autoload).
func (d *AcademicCalendarEventsDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		ACRelAcademicYear: {
			Domain:     "academic_year",
			Type:       vernon.RelBelongsTo,
			FK:         ACFieldAcademicYearID,
			LocalKey:   ACFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "start_date", "end_date"},
		},
	}
}

// Validate memvalidasi invariant domain academic_calendar_events sebelum write.
func (d *AcademicCalendarEventsDescriptor) Validate(data map[string]any) error {
	if err := acValidateRequired(data); err != nil {
		return err
	}
	return acValidateFormats(data)
}

// acValidateRequired memeriksa field wajib academic_calendar_events.
func acValidateRequired(data map[string]any) error {
	name, _ := data[ACFieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	eventType, _ := data[ACFieldEventType].(string)
	if eventType == "" {
		return errors.New("event_type wajib diisi")
	}
	if !validEventTypes[eventType] {
		return fmt.Errorf("event_type tidak valid: %q", eventType)
	}

	startDate, _ := data[ACFieldStartDate].(string)
	if startDate == "" {
		return errors.New("start_date wajib diisi")
	}

	endDate, _ := data[ACFieldEndDate].(string)
	if endDate == "" {
		return errors.New("end_date wajib diisi")
	}
	return nil
}

// acValidateFormats memeriksa format dan konsistensi field.
func acValidateFormats(data map[string]any) error {
	if err := acValidateDateRange(data); err != nil {
		return err
	}

	if err := acValidateSemester(data); err != nil {
		return err
	}

	return acValidateColorCode(data)
}

// acValidateDateRange memeriksa start_date <= end_date.
func acValidateDateRange(data map[string]any) error {
	startStr, _ := data[ACFieldStartDate].(string)
	endStr, _ := data[ACFieldEndDate].(string)

	start, err := time.Parse(dateFormat, startStr)
	if err != nil {
		return fmt.Errorf("start_date format tidak valid: %q (harus YYYY-MM-DD)", startStr)
	}

	end, err := time.Parse(dateFormat, endStr)
	if err != nil {
		return fmt.Errorf("end_date format tidak valid: %q (harus YYYY-MM-DD)", endStr)
	}

	if start.After(end) {
		return errors.New("start_date tidak boleh lebih besar dari end_date")
	}
	return nil
}

// acValidateSemester memeriksa semester jika diisi.
func acValidateSemester(data map[string]any) error {
	semester, _ := data[ACFieldSemester].(string)
	if semester != "" && !validACSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}
	return nil
}

// acValidateColorCode memeriksa color_code jika diisi.
func acValidateColorCode(data map[string]any) error {
	colorCode, _ := data[ACFieldColorCode].(string)
	if colorCode != "" && !colorCodeRegex.MatchString(colorCode) {
		return fmt.Errorf("color_code tidak valid: %q (harus format #RRGGBB)", colorCode)
	}
	return nil
}
