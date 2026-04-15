package academic_year_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/academic_year"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &academic_year.Descriptor{}
	if got := d.TableName(); got != "academic_years" {
		t.Errorf("TableName() = %q, want %q", got, "academic_years")
	}
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &academic_year.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestValidate_RejectsEmptyName(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name": "",
		"code": "2526",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject empty name")
	}
}

func TestValidate_RejectsEmptyCode(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name": "2025/2026",
		"code": "",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject empty code")
	}
}

func TestValidate_RejectsInvalidStatus(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name":   "2025/2026",
		"code":   "2526",
		"status": "archived",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}

func TestValidate_RejectsStartDateAfterEndDate(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name":       "2025/2026",
		"code":       "2526",
		"start_date": "2026-07-01",
		"end_date":   "2025-06-30",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject start_date >= end_date")
	}
}

func TestValidate_RejectsSemesterOverlap(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name":            "2025/2026",
		"code":            "2526",
		"start_date":      "2025-07-01",
		"end_date":        "2026-06-30",
		"semester1_start": "2025-07-14",
		"semester1_end":   "2026-02-15",
		"semester2_start": "2026-01-10",
		"semester2_end":   "2026-06-20",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject semester1_end > semester2_start")
	}
}

func TestValidate_AcceptsValidAcademicYear(t *testing.T) {
	d := &academic_year.Descriptor{}
	data := map[string]any{
		"name":            "2025/2026",
		"code":            "2526",
		"status":          "planning",
		"start_date":      "2025-07-14",
		"end_date":        "2026-06-20",
		"semester1_start": "2025-07-14",
		"semester1_end":   "2025-12-19",
		"semester2_start": "2026-01-05",
		"semester2_end":   "2026-06-20",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}
