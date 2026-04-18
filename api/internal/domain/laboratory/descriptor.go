// Package laboratory adalah domain Vernon untuk manajemen laboratorium.
//
// Terdiri dari 3 tabel: laboratories, lab_equipment, lab_usage_logs.
// laboratories tidak memiliki BelongsTo (root master).
// lab_equipment memiliki 1 BelongsTo autoload: lab.
// lab_usage_logs memiliki 4 BelongsTo autoload: lab, teacher, class_room, academic_year.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package laboratory

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Lab field constants ───────────────────────────────────────────────────────

const (
	LabFieldName              = "name"
	LabFieldType              = "type"
	LabFieldCapacity          = "capacity"
	LabFieldBuilding          = "building"
	LabFieldFloor             = "floor"
	LabFieldRoomNumber        = "room_number"
	LabFieldEquipmentCount    = "equipment_count"
	LabFieldSafetyRating      = "safety_rating"
	LabFieldLastInspection    = "last_inspection_date"
	LabFieldStatus            = "status"
)

// ── Equipment field constants ─────────────────────────────────────────────────

const (
	EqFieldLabID            = "lab_id"
	EqFieldName             = "name"
	EqFieldCategory         = "category"
	EqFieldBrand            = "brand"
	EqFieldModel            = "model"
	EqFieldSerialNumber     = "serial_number"
	EqFieldCondition        = "condition"
	EqFieldQuantity         = "quantity"
	EqFieldUnit             = "unit"
	EqFieldPurchaseDate     = "purchase_date"
	EqFieldPurchasePrice    = "purchase_price"
	EqFieldCalibrationDate  = "calibration_date"
	EqFieldNextCalibration  = "next_calibration"
)

// ── Usage log field constants ─────────────────────────────────────────────────

const (
	ULFieldLabID           = "lab_id"
	ULFieldTeacherID       = "teacher_id"
	ULFieldClassRoomID     = "class_room_id"
	ULFieldAcademicYearID  = "academic_year_id"
	ULFieldSubjectID       = "subject_id"
	ULFieldUsageDate       = "usage_date"
	ULFieldStartTime       = "start_time"
	ULFieldEndTime         = "end_time"
	ULFieldTopic           = "topic"
	ULFieldParticipantCount = "participant_count"
	ULFieldNotes           = "notes"
)

// ── Lab enum constants ────────────────────────────────────────────────────────

const (
	LabTypeIPA        = "ipa"
	LabTypeKomputer   = "komputer"
	LabTypeBahasa     = "bahasa"
	LabTypeMultimedia = "multimedia"
)

const (
	SafetyExcellent = "excellent"
	SafetyGood      = "good"
	SafetyFair      = "fair"
	SafetyPoor      = "poor"
)

const (
	LabStatusActive      = "active"
	LabStatusInactive    = "inactive"
	LabStatusMaintenance = "maintenance"
)

// ── Equipment enum constants ──────────────────────────────────────────────────

const (
	EqCatOptical    = "optical"
	EqCatElectronic = "electronic"
	EqCatMeasuring  = "measuring"
	EqCatChemical   = "chemical"
	EqCatBiological = "biological"
	EqCatSpecimen   = "specimen"
	EqCatTool       = "tool"
	EqCatSafety     = "safety"
	EqCatFurniture  = "furniture"
	EqCatComputer   = "computer"
	EqCatOther      = "other"
)

const (
	EqCondNew        = "new"
	EqCondGood       = "good"
	EqCondNeedsRepair = "needs_repair"
	EqCondDamaged    = "damaged"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelLab           = "lab"
	RelTeacher       = "teacher"
	RelClassRoom     = "class_room"
	RelAcademicYear  = "academic_year"
)

var validLabTypes = map[string]bool{
	LabTypeIPA: true, LabTypeKomputer: true, LabTypeBahasa: true, LabTypeMultimedia: true,
}

var validSafetyRatings = map[string]bool{
	SafetyExcellent: true, SafetyGood: true, SafetyFair: true, SafetyPoor: true,
}

var validLabStatuses = map[string]bool{
	LabStatusActive: true, LabStatusInactive: true, LabStatusMaintenance: true,
}

var validEquipmentCategories = map[string]bool{
	EqCatOptical: true, EqCatElectronic: true, EqCatMeasuring: true,
	EqCatChemical: true, EqCatBiological: true, EqCatSpecimen: true,
	EqCatTool: true, EqCatSafety: true, EqCatFurniture: true,
	EqCatComputer: true, EqCatOther: true,
}

var validEquipmentConditions = map[string]bool{
	EqCondNew: true, EqCondGood: true, EqCondNeedsRepair: true, EqCondDamaged: true,
}

// ── LabDescriptor ─────────────────────────────────────────────────────────────

// LabDescriptor mengimplementasi vernon.DomainDescriptor untuk laboratories.
type LabDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LabDescriptor) TableName() string { return "laboratories" }

// DefaultRels — laboratories tidak memiliki relasi (root master).
func (d *LabDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant laboratories.
func (d *LabDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, LabFieldName, "name"); err != nil {
		return err
	}
	return requireEnum(data, LabFieldType, "type", validLabTypes)
}

// ── EquipmentDescriptor ───────────────────────────────────────────────────────

// EquipmentDescriptor mengimplementasi vernon.DomainDescriptor untuk lab_equipment.
type EquipmentDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *EquipmentDescriptor) TableName() string { return "lab_equipment" }

// DefaultRels mendefinisikan relasi lab_equipment: lab.
func (d *EquipmentDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLab: {
			Domain:     "laboratories",
			Type:       vernon.RelBelongsTo,
			FK:         EqFieldLabID,
			LocalKey:   EqFieldLabID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "type", "room_number"},
		},
	}
}

// Validate memvalidasi invariant lab_equipment.
func (d *EquipmentDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, EqFieldLabID, "lab_id"); err != nil {
		return err
	}
	if err := requireString(data, EqFieldName, "name"); err != nil {
		return err
	}
	return requireEnum(data, EqFieldCategory, "category", validEquipmentCategories)
}

// ── UsageLogDescriptor ────────────────────────────────────────────────────────

// UsageLogDescriptor mengimplementasi vernon.DomainDescriptor untuk lab_usage_logs.
type UsageLogDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *UsageLogDescriptor) TableName() string { return "lab_usage_logs" }

// DefaultRels mendefinisikan relasi lab_usage_logs: lab + teacher + class_room + academic_year.
func (d *UsageLogDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLab: {
			Domain:     "laboratories",
			Type:       vernon.RelBelongsTo,
			FK:         ULFieldLabID,
			LocalKey:   ULFieldLabID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "type"},
		},
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         ULFieldTeacherID,
			LocalKey:   ULFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		RelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         ULFieldClassRoomID,
			LocalKey:   ULFieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         ULFieldAcademicYearID,
			LocalKey:   ULFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant lab_usage_logs.
func (d *UsageLogDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ULFieldLabID, "lab_id"); err != nil {
		return err
	}
	if err := requireString(data, ULFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	return requireString(data, ULFieldUsageDate, "usage_date")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func requireEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}
