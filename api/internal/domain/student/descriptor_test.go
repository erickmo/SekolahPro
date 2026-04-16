package student_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student.Descriptor{}
	if got := d.TableName(); got != "students" {
		t.Errorf("TableName() = %q, want %q", got, "students")
	}
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &student.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name":  "Ahmad Rizky",
		"nis":        "12345",
		"nisn":       "0012345678",
		"gender":     "L",
		"status":     "active",
		"blood_type": "O",
		"religion":   "islam",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_FullName(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"gender": "L",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing full_name")
	}
}

func TestDescriptor_Validate_MissingRequired_Gender(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name": "Ahmad Rizky",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing gender")
	}
}

func TestDescriptor_Validate_InvalidEnums_Gender(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name": "Ahmad Rizky",
		"gender":    "X",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid gender")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name": "Ahmad Rizky",
		"gender":    "L",
		"status":    "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}

func TestDescriptor_Validate_InvalidEnums_BloodType(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name":  "Ahmad Rizky",
		"gender":     "L",
		"blood_type": "Z",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid blood_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_Religion(t *testing.T) {
	d := &student.Descriptor{}
	data := map[string]any{
		"full_name": "Ahmad Rizky",
		"gender":    "L",
		"religion":  "atheism",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid religion")
	}
}
