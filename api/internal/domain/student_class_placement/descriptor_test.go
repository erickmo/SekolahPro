package student_class_placement_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student_class_placement"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	if got := d.TableName(); got != "student_class_placements" {
		t.Errorf("TableName() = %q, want %q", got, "student_class_placements")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 3 {
		t.Errorf("DefaultRels() length = %d, want 3", len(rels))
	}
	for _, name := range []string{"student", "class_room", "academic_year"} {
		if _, ok := rels[name]; !ok {
			t.Errorf("DefaultRels() missing %s relation", name)
		}
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"student_id":       "018f0000-0000-7000-8000-000000000001",
		"class_room_id":    "018f0000-0000-7000-8000-000000000002",
		"academic_year_id": "018f0000-0000-7000-8000-000000000003",
		"semester":         "1",
		"status":           "active",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_StudentID(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"class_room_id":    "018f0000-0000-7000-8000-000000000002",
		"academic_year_id": "018f0000-0000-7000-8000-000000000003",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing student_id")
	}
}

func TestDescriptor_Validate_MissingRequired_ClassRoomID(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"student_id":       "018f0000-0000-7000-8000-000000000001",
		"academic_year_id": "018f0000-0000-7000-8000-000000000003",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing class_room_id")
	}
}

func TestDescriptor_Validate_MissingRequired_AcademicYearID(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"student_id":    "018f0000-0000-7000-8000-000000000001",
		"class_room_id": "018f0000-0000-7000-8000-000000000002",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing academic_year_id")
	}
}

func TestDescriptor_Validate_InvalidEnums_Semester(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"student_id":       "018f0000-0000-7000-8000-000000000001",
		"class_room_id":    "018f0000-0000-7000-8000-000000000002",
		"academic_year_id": "018f0000-0000-7000-8000-000000000003",
		"semester":         "3",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid semester")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &student_class_placement.Descriptor{}
	data := map[string]any{
		"student_id":       "018f0000-0000-7000-8000-000000000001",
		"class_room_id":    "018f0000-0000-7000-8000-000000000002",
		"academic_year_id": "018f0000-0000-7000-8000-000000000003",
		"status":           "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}
