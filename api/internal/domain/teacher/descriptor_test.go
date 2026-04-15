package teacher_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/teacher"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &teacher.Descriptor{}
	if got := d.TableName(); got != "teachers" {
		t.Errorf("TableName() = %q, want %q", got, "teachers")
	}
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &teacher.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestValidate_RejectsEmptyFullName(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "",
		"gender":        "L",
		"employee_type": "pns",
		"role":          "guru_mapel",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject empty full_name")
	}
}

func TestValidate_RejectsInvalidRole(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "Pak Budi",
		"gender":        "L",
		"employee_type": "pns",
		"role":          "janitor",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid role")
	}
}

func TestValidate_RejectsInvalidEmployeeType(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "Bu Siti",
		"gender":        "P",
		"employee_type": "freelance",
		"role":          "guru_mapel",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid employee_type")
	}
}

func TestValidate_RejectsInvalidGender(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "Test",
		"gender":        "X",
		"employee_type": "pns",
		"role":          "guru_mapel",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid gender")
	}
}

func TestValidate_RejectsInvalidStatus(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "Bu Ani",
		"gender":        "P",
		"employee_type": "honorer",
		"role":          "admin_tu",
		"status":        "fired",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}

func TestValidate_AcceptsValidTeacher(t *testing.T) {
	d := &teacher.Descriptor{}
	data := map[string]any{
		"full_name":     "Pak Budi Santoso",
		"gender":        "L",
		"employee_type": "pns",
		"role":          "guru_mapel",
		"status":        "active",
		"nip":           "198501012010011001",
		"religion":      "islam",
		"join_date":     "2010-01-01",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}
