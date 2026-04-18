// Package asset adalah domain Vernon untuk manajemen aset & inventaris.
//
// Terdiri dari 2 tabel: assets, asset_maintenances.
// assets memiliki 1 BelongsTo autoload opsional: room (class_rooms).
// asset_maintenances memiliki 1 BelongsTo autoload: asset.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package asset

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Asset field constants ─────────────────────────────────────────────────────

const (
	FieldAssetCode         = "asset_code"
	FieldName              = "name"
	FieldCategory          = "category"
	FieldDescription       = "description"
	FieldAcquisitionDate   = "acquisition_date"
	FieldAcquisitionCost   = "acquisition_cost"
	FieldOwnershipType     = "ownership_type"
	FieldCondition         = "condition"
	FieldLocationRoomID    = "location_room_id"
	FieldResponsiblePerson = "responsible_person"
	FieldVendor            = "vendor"
	FieldWarrantyExpiry    = "warranty_expiry"
	FieldDepreciationMethod = "depreciation_method"
	FieldUsefulLifeYears   = "useful_life_years"
	FieldBookValue         = "book_value"
	FieldLifecycleStatus   = "lifecycle_status"
	FieldDisposalDate      = "disposal_date"
)

// ── Maintenance field constants ───────────────────────────────────────────────

const (
	MtFieldAssetID        = "asset_id"
	MtFieldType           = "type"
	MtFieldDescription    = "description"
	MtFieldStartDate      = "start_date"
	MtFieldCompletionDate = "completion_date"
	MtFieldCost           = "cost"
	MtFieldVendor         = "vendor"
	MtFieldTechnician     = "technician"
	MtFieldStatus         = "status"
	MtFieldNotes          = "notes"
)

// ── Asset enum constants ──────────────────────────────────────────────────────

const (
	CatTanah     = "tanah"
	CatBangunan  = "bangunan"
	CatMesin     = "mesin"
	CatKendaraan = "kendaraan"
	CatPeralatan = "peralatan"
	CatLainnya   = "lainnya"
)

const (
	OwnOwned   = "owned"
	OwnLeased  = "leased"
	OwnDonated = "donated"
)

const (
	CondNew     = "new"
	CondGood    = "good"
	CondFair    = "fair"
	CondPoor    = "poor"
	CondDamaged = "damaged"
)

const (
	LSInUse       = "in_use"
	LSIdle        = "idle"
	LSMaintenance = "maintenance"
	LSDisposed    = "disposed"
	LSLost        = "lost"
	LSTransferred = "transferred"
)

// ── Maintenance enum constants ────────────────────────────────────────────────

const (
	MtPreventive  = "preventive"
	MtCorrective  = "corrective"
	MtEmergency   = "emergency"
	MtCalibration = "calibration"
	MtInspection  = "inspection"
)

const (
	MtScheduled   = "scheduled"
	MtInProgress  = "in_progress"
	MtCompleted   = "completed"
	MtCancelled   = "cancelled"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelRoom  = "room"
	RelAsset = "asset"
)

var validAssetCategories = map[string]bool{
	CatTanah: true, CatBangunan: true, CatMesin: true,
	CatKendaraan: true, CatPeralatan: true, CatLainnya: true,
}

var validOwnershipTypes = map[string]bool{
	OwnOwned: true, OwnLeased: true, OwnDonated: true,
}

var validAssetConditions = map[string]bool{
	CondNew: true, CondGood: true, CondFair: true, CondPoor: true, CondDamaged: true,
}

var validLifecycleStatuses = map[string]bool{
	LSInUse: true, LSIdle: true, LSMaintenance: true,
	LSDisposed: true, LSLost: true, LSTransferred: true,
}

var validMaintenanceTypes = map[string]bool{
	MtPreventive: true, MtCorrective: true, MtEmergency: true,
	MtCalibration: true, MtInspection: true,
}

var validMaintenanceStatuses = map[string]bool{
	MtScheduled: true, MtInProgress: true, MtCompleted: true, MtCancelled: true,
}

// ── AssetDescriptor ───────────────────────────────────────────────────────────

// AssetDescriptor mengimplementasi vernon.DomainDescriptor untuk assets.
type AssetDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *AssetDescriptor) TableName() string { return "assets" }

// DefaultRels mendefinisikan relasi assets: room (opsional).
func (d *AssetDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldLocationRoomID,
			LocalKey:   FieldLocationRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "building"},
		},
	}
}

// Validate memvalidasi invariant assets.
func (d *AssetDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldAssetCode, "asset_code"); err != nil {
		return err
	}
	if err := requireString(data, FieldName, "name"); err != nil {
		return err
	}
	return requireEnum(data, FieldCategory, "category", validAssetCategories)
}

// ── MaintenanceDescriptor ─────────────────────────────────────────────────────

// MaintenanceDescriptor mengimplementasi vernon.DomainDescriptor untuk asset_maintenances.
type MaintenanceDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *MaintenanceDescriptor) TableName() string { return "asset_maintenances" }

// DefaultRels mendefinisikan relasi asset_maintenances: asset.
func (d *MaintenanceDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelAsset: {
			Domain:     "assets",
			Type:       vernon.RelBelongsTo,
			FK:         MtFieldAssetID,
			LocalKey:   MtFieldAssetID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"asset_code", "name", "condition"},
		},
	}
}

// Validate memvalidasi invariant asset_maintenances.
func (d *MaintenanceDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, MtFieldAssetID, "asset_id"); err != nil {
		return err
	}
	if err := requireEnum(data, MtFieldType, "type", validMaintenanceTypes); err != nil {
		return err
	}
	if err := requireString(data, MtFieldDescription, "description"); err != nil {
		return err
	}
	return requireString(data, MtFieldStartDate, "start_date")
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
