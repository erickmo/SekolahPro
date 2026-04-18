// Package facility_booking adalah domain Vernon untuk booking ruangan & fasilitas.
//
// Terdiri dari 2 tabel: facilities, facility_bookings.
// facilities tidak memiliki BelongsTo (root master).
// facility_bookings memiliki 2 BelongsTo autoload: facility, academic_year.
// requester adalah polymorphic (5 tipe) dan TIDAK di-autoload.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package facility_booking

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Facility field constants ──────────────────────────────────────────────────

const (
	FFieldName             = "name"
	FFieldType             = "type"
	FFieldBuilding         = "building"
	FFieldFloor            = "floor"
	FFieldRoomNumber       = "room_number"
	FFieldCapacity         = "capacity"
	FFieldIsBookable       = "is_bookable"
	FFieldRequiresApproval = "requires_approval"
	FFieldStatus           = "status"
)

// ── Booking field constants ───────────────────────────────────────────────────

const (
	BFieldFacilityID       = "facility_id"
	BFieldRequesterType    = "requester_type"
	BFieldRequesterID      = "requester_id"
	BFieldAcademicYearID   = "academic_year_id"
	BFieldBookingDate      = "booking_date"
	BFieldStartTime        = "start_time"
	BFieldEndTime          = "end_time"
	BFieldPurpose          = "purpose"
	BFieldStatus           = "status"
	BFieldRecurrence       = "recurrence_pattern"
	BFieldNotes            = "notes"
)

// ── Facility enum constants ───────────────────────────────────────────────────

const (
	FTypeClassroom    = "classroom"
	FTypeLab          = "lab"
	FTypeLibrary      = "library"
	FTypeHall         = "hall"
	FTypeMeetingRoom  = "meeting_room"
	FTypeSportsField  = "sports_field"
	FTypeMosque       = "mosque"
	FTypeOther        = "other"
)

const (
	FStatusActive      = "active"
	FStatusInactive    = "inactive"
	FStatusMaintenance = "maintenance"
)

// ── Booking enum constants ────────────────────────────────────────────────────

const (
	BReqTeacher  = "teacher"
	BReqStudent  = "student"
	BReqStaff    = "staff"
	BReqExternal = "external"
	BReqAdmin    = "admin"
)

const (
	BStatusPending   = "pending"
	BStatusApproved  = "approved"
	BStatusRejected  = "rejected"
	BStatusCancelled = "cancelled"
	BStatusCompleted = "completed"
)

const (
	BRecDaily   = "daily"
	BRecWeekly  = "weekly"
	BRecMonthly = "monthly"
	BRecNone    = "none"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelFacility      = "facility"
	RelAcademicYear  = "academic_year"
)

var validFacilityTypes = map[string]bool{
	FTypeClassroom: true, FTypeLab: true, FTypeLibrary: true,
	FTypeHall: true, FTypeMeetingRoom: true, FTypeSportsField: true,
	FTypeMosque: true, FTypeOther: true,
}

var validFacilityStatuses = map[string]bool{
	FStatusActive: true, FStatusInactive: true, FStatusMaintenance: true,
}

var validRequesterTypes = map[string]bool{
	BReqTeacher: true, BReqStudent: true, BReqStaff: true,
	BReqExternal: true, BReqAdmin: true,
}

var validBookingStatuses = map[string]bool{
	BStatusPending: true, BStatusApproved: true, BStatusRejected: true,
	BStatusCancelled: true, BStatusCompleted: true,
}

// ── FacilityDescriptor ────────────────────────────────────────────────────────

// FacilityDescriptor mengimplementasi vernon.DomainDescriptor untuk facilities.
type FacilityDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *FacilityDescriptor) TableName() string { return "facilities" }

// DefaultRels — facilities tidak memiliki relasi (root master).
func (d *FacilityDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant facilities.
func (d *FacilityDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FFieldName, "name"); err != nil {
		return err
	}
	return requireEnum(data, FFieldType, "type", validFacilityTypes)
}

// ── BookingDescriptor ─────────────────────────────────────────────────────────

// BookingDescriptor mengimplementasi vernon.DomainDescriptor untuk facility_bookings.
type BookingDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *BookingDescriptor) TableName() string { return "facility_bookings" }

// DefaultRels mendefinisikan relasi facility_bookings: facility + academic_year.
// requester adalah polymorphic dan TIDAK di-autoload.
func (d *BookingDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelFacility: {
			Domain:     "facilities",
			Type:       vernon.RelBelongsTo,
			FK:         BFieldFacilityID,
			LocalKey:   BFieldFacilityID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "type", "building", "room_number"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         BFieldAcademicYearID,
			LocalKey:   BFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant facility_bookings.
func (d *BookingDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, BFieldFacilityID, "facility_id"); err != nil {
		return err
	}
	if err := requireEnum(data, BFieldRequesterType, "requester_type", validRequesterTypes); err != nil {
		return err
	}
	if err := requireString(data, BFieldRequesterID, "requester_id"); err != nil {
		return err
	}
	if err := requireString(data, BFieldBookingDate, "booking_date"); err != nil {
		return err
	}
	return requireString(data, BFieldPurpose, "purpose")
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
