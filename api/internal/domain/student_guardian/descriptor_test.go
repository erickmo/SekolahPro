package student_guardian_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student_guardian"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student_guardian.Descriptor{}
	if got := d.TableName(); got != "student_guardians" {
		t.Errorf("TableName() = %q, want %q", got, "student_guardians")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &student_guardian.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 1 {
		t.Errorf("DefaultRels() length = %d, want 1", len(rels))
	}
	if _, ok := rels["student"]; !ok {
		t.Error("DefaultRels() missing student relation")
	}
}

func TestDescriptor_Rels_StudentAutoload(t *testing.T) {
	d := &student_guardian.Descriptor{}
	rels := d.DefaultRels()
	student := rels["student"]
	if !student.IsAutoload {
		t.Error("student relation should be autoload")
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student_guardian.Descriptor{}
	data := map[string]any{
		"student_id":    "018f0000-0000-7000-8000-000000000001",
		"guardian_type": "ayah",
		"full_name":     "Budi Santoso",
		"phone":         "081234567890",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_StudentID(t *testing.T) {
	d := &student_guardian.Descriptor{}
	data := map[string]any{
		"guardian_type": "ayah",
		"full_name":     "Budi Santoso",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing student_id")
	}
}

func TestDescriptor_Validate_MissingRequired_GuardianType(t *testing.T) {
	d := &student_guardian.Descriptor{}
	data := map[string]any{
		"student_id": "018f0000-0000-7000-8000-000000000001",
		"full_name":  "Budi Santoso",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing guardian_type")
	}
}

func TestDescriptor_Validate_MissingRequired_FullName(t *testing.T) {
	d := &student_guardian.Descriptor{}
	data := map[string]any{
		"student_id":    "018f0000-0000-7000-8000-000000000001",
		"guardian_type": "ayah",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing full_name")
	}
}

func TestDescriptor_Validate_InvalidEnums_GuardianType(t *testing.T) {
	d := &student_guardian.Descriptor{}
	data := map[string]any{
		"student_id":    "018f0000-0000-7000-8000-000000000001",
		"guardian_type": "paman",
		"full_name":     "Budi Santoso",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid guardian_type")
	}
}
