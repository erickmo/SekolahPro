package student_admission_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student_admission"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student_admission.Descriptor{}
	if got := d.TableName(); got != "student_admissions" {
		t.Errorf("TableName() = %q, want %q", got, "student_admissions")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &student_admission.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 2 {
		t.Errorf("DefaultRels() length = %d, want 2", len(rels))
	}
	for _, name := range []string{"student", "academic_year"} {
		if _, ok := rels[name]; !ok {
			t.Errorf("DefaultRels() missing %s relation", name)
		}
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"registration_number": "PPDB-2025-001",
		"student_id":          "018f0000-0000-7000-8000-000000000001",
		"academic_year_id":    "018f0000-0000-7000-8000-000000000002",
		"admission_type":      "regular",
		"status":              "pending",
		"test_score":          float64(85.5),
		"interview_score":     float64(90),
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_RegistrationNumber(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"admission_type": "regular",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing registration_number")
	}
}

func TestDescriptor_Validate_MissingRequired_AdmissionType(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"registration_number": "PPDB-2025-001",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing admission_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_AdmissionType(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"registration_number": "PPDB-2025-001",
		"admission_type":      "beasiswa",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid admission_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"registration_number": "PPDB-2025-001",
		"admission_type":      "regular",
		"status":              "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}

func TestDescriptor_Validate_NegativeScore(t *testing.T) {
	d := &student_admission.Descriptor{}
	data := map[string]any{
		"registration_number": "PPDB-2025-001",
		"admission_type":      "regular",
		"test_score":          float64(-10),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject negative test_score")
	}
}
