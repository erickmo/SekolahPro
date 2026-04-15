package class_room_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/class_room"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &class_room.Descriptor{}
	if got := d.TableName(); got != "class_rooms" {
		t.Errorf("TableName() = %q, want %q", got, "class_rooms")
	}
}

func TestDescriptor_DefaultRels_Returns2(t *testing.T) {
	d := &class_room.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 2 {
		t.Errorf("DefaultRels() length = %d, want 2", len(rels))
	}
	if _, ok := rels["academic_year"]; !ok {
		t.Error("DefaultRels() missing academic_year relation")
	}
	if _, ok := rels["homeroom_teacher"]; !ok {
		t.Error("DefaultRels() missing homeroom_teacher relation")
	}
}

func TestDescriptor_Rels_AcademicYearAutoload(t *testing.T) {
	d := &class_room.Descriptor{}
	rels := d.DefaultRels()
	ay := rels["academic_year"]
	if !ay.IsAutoload {
		t.Error("academic_year relation should be autoload")
	}
	if ay.Domain != "academic_years" {
		t.Errorf("academic_year.Domain = %q, want %q", ay.Domain, "academic_years")
	}
}

func TestValidate_RejectsEmptyName(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "",
		"grade_level":      "7",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject empty name")
	}
}

func TestValidate_RejectsInvalidGradeLevel(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "XIII-A",
		"grade_level":      "13",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid grade_level")
	}
}

func TestValidate_RejectsZeroCapacity(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "VII-A",
		"grade_level":      "7",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
		"capacity":         float64(0),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject capacity <= 0")
	}
}

func TestValidate_RejectsInvalidMajor(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "XI-TEKNIK-1",
		"grade_level":      "11",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
		"major":            "teknik",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid major")
	}
}

func TestValidate_AcceptsValidClassRoom(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "VII-A",
		"grade_level":      "7",
		"parallel_id":      "A",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
		"capacity":         float64(36),
		"current_count":    float64(0),
		"is_active":        true,
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestValidate_AcceptsValidSMAWithMajor(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":             "XI-IPA-1",
		"grade_level":      "11",
		"academic_year_id": "018f0000-0000-7000-8000-000000000001",
		"major":            "ipa",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept SMA class with major, got: %v", err)
	}
}

func TestValidate_RejectsMissingAcademicYearID(t *testing.T) {
	d := &class_room.Descriptor{}
	data := map[string]any{
		"name":        "VII-A",
		"grade_level": "7",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing academic_year_id")
	}
}
