// Package teacher_workload adalah domain Vernon untuk beban mengajar guru.
//
// Terdiri dari 2 tabel: teacher_workloads dan teacher_workload_items.
// teacher_workloads memiliki 2 BelongsTo autoload: teacher dan academic_year.
// teacher_workload_items memiliki 2 BelongsTo autoload: workload dan teacher.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher_workload

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Workload field name constants ──────────────────────────────────────────────

const (
	FieldTeacherID         = "teacher_id"
	FieldAcademicYearID    = "academic_year_id"
	FieldSemester          = "semester"
	FieldTeachingHours     = "teaching_hours"
	FieldAdditionalHours   = "additional_hours"
	FieldTotalHours        = "total_hours"
	FieldMinimumRequired   = "minimum_required"
	FieldFulfillmentStatus = "fulfillment_status"
)

// ── Workload item field name constants ─────────────────────────────────────────

const (
	FieldWorkloadID   = "workload_id"
	FieldItemType     = "item_type"
	FieldDescription  = "description"
	FieldHoursPerWeek = "hours_per_week"
)

// ── Semester constants ─────────────────────────────────────────────────────────

const (
	SemesterGanjil = "ganjil"
	SemesterGenap  = "genap"
)

// ── Fulfillment status constants ───────────────────────────────────────────────

const (
	FulfillmentKurang    = "kurang"
	FulfillmentTerpenuhi = "terpenuhi"
	FulfillmentLebih     = "lebih"
)

// ── Item type constants ────────────────────────────────────────────────────────

const (
	ItemTypeMengajar      = "mengajar"
	ItemTypeWaliKelas     = "wali_kelas"
	ItemTypePembinaEkskul = "pembina_ekskul"
	ItemTypeGuruBK        = "guru_bk"
	ItemTypeKepalaSekolah = "kepala_sekolah"
	ItemTypeWakilKepsek   = "wakil_kepsek"
	ItemTypeKoordinator   = "koordinator"
	ItemTypePanitia       = "panitia"
	ItemTypeTugasTambahan = "tugas_tambahan"
)

// ── Relation name constants ────────────────────────────────────────────────────

const (
	RelTeacher      = "teacher"
	RelAcademicYear = "academic_year"
	RelWorkload     = "workload"
)

var validSemesters = map[string]bool{
	SemesterGanjil: true, SemesterGenap: true,
}

var validFulfillmentStatuses = map[string]bool{
	FulfillmentKurang: true, FulfillmentTerpenuhi: true, FulfillmentLebih: true,
}

var validItemTypes = map[string]bool{
	ItemTypeMengajar: true, ItemTypeWaliKelas: true,
	ItemTypePembinaEkskul: true, ItemTypeGuruBK: true,
	ItemTypeKepalaSekolah: true, ItemTypeWakilKepsek: true,
	ItemTypeKoordinator: true, ItemTypePanitia: true,
	ItemTypeTugasTambahan: true,
}

// ── Descriptor (teacher_workloads) ─────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk teacher_workloads.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "teacher_workloads" }

// DefaultRels mendefinisikan relasi teacher_workloads: teacher + academic_year.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherID,
			LocalKey:   FieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type", "role"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant teacher_workloads sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}

	semester, _ := data[FieldSemester].(string)
	if semester == "" {
		return fmt.Errorf("semester wajib diisi")
	}
	if !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q", semester)
	}
	return nil
}

// ── ItemDescriptor (teacher_workload_items) ────────────────────────────────────

// ItemDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_workload_items.
type ItemDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk item beban mengajar.
func (d *ItemDescriptor) TableName() string { return "teacher_workload_items" }

// DefaultRels mendefinisikan relasi teacher_workload_items: workload + teacher.
func (d *ItemDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelWorkload: {
			Domain:     "teacher_workloads",
			Type:       vernon.RelBelongsTo,
			FK:         FieldWorkloadID,
			LocalKey:   FieldWorkloadID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"total_hours", "is_fulfilled"},
		},
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherID,
			LocalKey:   FieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name"},
		},
	}
}

// Validate memvalidasi invariant teacher_workload_items sebelum write.
func (d *ItemDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldWorkloadID, "workload_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}

	itemType, _ := data[FieldItemType].(string)
	if itemType == "" {
		return fmt.Errorf("item_type wajib diisi")
	}
	if !validItemTypes[itemType] {
		return fmt.Errorf("item_type tidak valid: %q", itemType)
	}

	if err := requireString(data, FieldDescription, "description"); err != nil {
		return err
	}

	hoursRaw, ok := data[FieldHoursPerWeek]
	if !ok {
		return fmt.Errorf("hours_per_week wajib diisi")
	}
	hours, isNum := hoursRaw.(float64)
	if !isNum || hours < 0 {
		return fmt.Errorf("hours_per_week harus berupa angka >= 0")
	}
	return nil
}

// ── shared validation helpers ──────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}
