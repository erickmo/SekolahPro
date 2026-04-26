// Package toko_meal_plan adalah domain Vernon untuk meal plan asrama.
//
// Autoloads: nasabah, rekening.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_meal_plan

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldNasabahID = "nasabah_id"
	FieldRekeningID = "rekening_id"
	FieldPlanName  = "plan_name"
	FieldMealType  = "meal_type"
	FieldFrequency = "frequency"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	MealBreakfast = "breakfast"
	MealLunch     = "lunch"
	MealDinner    = "dinner"
	MealSnack     = "snack"
)

const (
	FreqDaily   = "daily"
	FreqWeekly  = "weekly"
	FreqMonthly = "monthly"
)

var validMealTypes = map[string]bool{
	MealBreakfast: true, MealLunch: true,
	MealDinner: true, MealSnack: true,
}

var validFrequencies = map[string]bool{
	FreqDaily: true, FreqWeekly: true, FreqMonthly: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelNasabah  = "nasabah"
	RelRekening = "rekening"
)

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_meal_plan.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_meal_plan" }

// DefaultRels mendefinisikan relasi toko_meal_plan: nasabah + rekening.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         FieldRekeningID,
			LocalKey:   FieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"account_number", "balance"},
		},
	}
}

// Validate memvalidasi invariant toko_meal_plan.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldNasabahID, "nasabah"); err != nil {
		return err
	}
	if err := requireString(data, FieldRekeningID, "rekening"); err != nil {
		return err
	}
	if err := requireString(data, FieldPlanName, "nama plan"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldMealType, "tipe makanan", validMealTypes); err != nil {
		return err
	}
	return requireEnum(data, FieldFrequency, "frekuensi", validFrequencies)
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
