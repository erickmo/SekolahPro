package academic_calendar_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/academic_calendar"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

func TestAcademicCalendarEventsDescriptor_TableName(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	if got := d.TableName(); got != "academic_calendar_events" {
		t.Errorf("TableName() = %q, want %q", got, "academic_calendar_events")
	}
}

func TestAcademicCalendarEventsDescriptor_DefaultRels(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	rels := d.DefaultRels()

	if len(rels) != 1 {
		t.Fatalf("DefaultRels() length = %d, want 1", len(rels))
	}

	rel, ok := rels["academic_year"]
	if !ok {
		t.Fatal(`DefaultRels() missing "academic_year" relation`)
	}

	if rel.Type != vernon.RelBelongsTo {
		t.Errorf("academic_year type = %q, want %q", rel.Type, vernon.RelBelongsTo)
	}
	if !rel.IsAutoload {
		t.Error("academic_year IsAutoload should be true")
	}
	if rel.FK != "academic_year_id" {
		t.Errorf("academic_year FK = %q, want %q", rel.FK, "academic_year_id")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_Valid(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":            "Libur Hari Raya Idul Fitri",
		"event_type":      "holiday",
		"start_date":      "2026-03-20",
		"end_date":        "2026-03-25",
		"semester":        "ganjil",
		"color_code":      "#FF5733",
		"academic_year_id": "uuid-year-001",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_ValidWithoutOptionals(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Ujian Tengah Semester",
		"event_type": "exam_period",
		"start_date": "2026-09-15",
		"end_date":   "2026-09-22",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data without optionals, got: %v", err)
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_SameDayEvent(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Hari Pendidikan Nasional",
		"event_type": "school_event",
		"start_date": "2026-05-02",
		"end_date":   "2026-05-02",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept same-day event, got: %v", err)
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_MissingName(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"event_type": "holiday",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing name")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_MissingEventType(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing event_type")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_InvalidEventType(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "invalid_type",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid event_type")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_MissingStartDate(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"end_date":   "2026-03-25",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing start_date")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_MissingEndDate(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "2026-03-20",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing end_date")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_StartDateAfterEndDate(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "2026-03-25",
		"end_date":   "2026-03-20",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject start_date > end_date")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_InvalidStartDate(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "not-a-date",
		"end_date":   "2026-03-25",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid start_date format")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
		"semester":   "semester1",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid semester")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_InvalidColorCode(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
		"color_code": "red",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid color_code")
	}
}

func TestAcademicCalendarEventsDescriptor_Validate_ColorCodeMissingHash(t *testing.T) {
	d := &academic_calendar.AcademicCalendarEventsDescriptor{}
	data := map[string]any{
		"name":       "Test Event",
		"event_type": "holiday",
		"start_date": "2026-03-20",
		"end_date":   "2026-03-25",
		"color_code": "FF5733",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject color_code without #")
	}
}
