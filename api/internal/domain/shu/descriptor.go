// Package shu adalah domain Vernon untuk SHU (Sisa Hasil Usaha) koperasi.
//
// Terdiri dari 2 Vernon descriptor:
//   - SHUPeriode: header periode SHU, tanpa BelongsTo selain tenant.
//   - SHUAnggota: detail per anggota, BelongsTo shu_periode, nasabah, rekening.
//
// shu_config bukan Vernon domain — hanya regular config table.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package shu

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// SHUPeriode
// ============================================================================

// SHUPeriode field name constants.
const (
	SPFieldTahunBuku        = "tahun_buku"
	SPFieldPeriodStart      = "period_start"
	SPFieldPeriodEnd        = "period_end"
	SPFieldTotalPendapatan  = "total_pendapatan"
	SPFieldTotalBeban       = "total_beban"
	SPFieldSHUBruto         = "shu_bruto"
	SPFieldSHUNeto          = "shu_neto"
	SPFieldDistributionCfg  = "distribution_config"
	SPFieldStatus           = "status"
)

// SHUPeriode status constants.
const (
	SPStatusCalculated  = "calculated"
	SPStatusReviewed    = "reviewed"
	SPStatusApproved    = "approved"
	SPStatusDistributed = "distributed"
)

var validSPStatuses = map[string]bool{
	SPStatusCalculated: true, SPStatusReviewed: true,
	SPStatusApproved: true, SPStatusDistributed: true,
}

// SHUPeriodeDescriptor mengimplementasi vernon.DomainDescriptor untuk shu_periode.
type SHUPeriodeDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *SHUPeriodeDescriptor) TableName() string { return "shu_periode" }

// DefaultRels — SHUPeriode tidak memiliki BelongsTo selain tenant.
func (d *SHUPeriodeDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain shu_periode sebelum write.
func (d *SHUPeriodeDescriptor) Validate(data map[string]any) error {
	if err := validateSPRequired(data); err != nil {
		return err
	}
	return validateSPEnums(data)
}

func validateSPRequired(data map[string]any) error {
	tb, ok := data[SPFieldTahunBuku].(float64)
	if !ok || tb < 2000 {
		return fmt.Errorf("tahun_buku wajib diisi dan harus >= 2000")
	}

	ps, _ := data[SPFieldPeriodStart].(string)
	if ps == "" {
		return fmt.Errorf("period_start wajib diisi")
	}

	pe, _ := data[SPFieldPeriodEnd].(string)
	if pe == "" {
		return fmt.Errorf("period_end wajib diisi")
	}

	tp, ok := data[SPFieldTotalPendapatan].(float64)
	if !ok {
		return fmt.Errorf("total_pendapatan wajib diisi dan harus berupa angka")
	}
	if tp < 0 {
		return fmt.Errorf("total_pendapatan tidak boleh negatif")
	}

	tb2, ok := data[SPFieldTotalBeban].(float64)
	if !ok {
		return fmt.Errorf("total_beban wajib diisi dan harus berupa angka")
	}
	if tb2 < 0 {
		return fmt.Errorf("total_beban tidak boleh negatif")
	}

	sb, ok := data[SPFieldSHUBruto].(float64)
	if !ok {
		return fmt.Errorf("shu_bruto wajib diisi dan harus berupa angka")
	}
	if sb < 0 {
		return fmt.Errorf("shu_bruto tidak boleh negatif")
	}

	sn, ok := data[SPFieldSHUNeto].(float64)
	if !ok {
		return fmt.Errorf("shu_neto wajib diisi dan harus berupa angka")
	}
	if sn < 0 {
		return fmt.Errorf("shu_neto tidak boleh negatif")
	}
	return nil
}

func validateSPEnums(data map[string]any) error {
	status, _ := data[SPFieldStatus].(string)
	if status != "" && !validSPStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ============================================================================
// SHUAnggota
// ============================================================================

// SHUAnggota field name constants.
const (
	SAFieldSHUPeriodeID      = "shu_periode_id"
	SAFieldNasabahID         = "nasabah_id"
	SAFieldAvgSimpanan       = "avg_simpanan"
	SAFieldTotalTransaksi    = "total_transaksi"
	SAFieldActiveDays        = "active_days"
	SAFieldJasaModal         = "jasa_modal"
	SAFieldJasaUsaha         = "jasa_usaha"
	SAFieldTotalSHU          = "total_shu"
	SAFieldDistMethod        = "distribution_method"
	SAFieldTargetRekeningID  = "target_rekening_id"
)

// SHUAnggota distribution method constants.
const (
	DistMethodCreditTabungan  = "credit_tabungan"
	DistMethodSeparatePayout  = "separate_payout"
	DistMethodPending         = "pending"
)

// SHUAnggota relation name constants.
const (
	RelSHUPeriode      = "shu_periode"
	RelNasabah         = "nasabah"
	RelTargetRekening  = "target_rekening"
)

var validDistMethods = map[string]bool{
	DistMethodCreditTabungan: true, DistMethodSeparatePayout: true,
	DistMethodPending: true,
}

// SHUAnggotaDescriptor mengimplementasi vernon.DomainDescriptor untuk shu_anggota.
type SHUAnggotaDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *SHUAnggotaDescriptor) TableName() string { return "shu_anggota" }

// DefaultRels mendefinisikan relasi domain shu_anggota.
// BelongsTo shu_periode (autoload), nasabah (autoload), target_rekening (autoload).
func (d *SHUAnggotaDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelSHUPeriode: {
			Domain:     "shu_periode",
			Type:       vernon.RelBelongsTo,
			FK:         SAFieldSHUPeriodeID,
			LocalKey:   SAFieldSHUPeriodeID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"tahun_buku", "status"},
		},
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         SAFieldNasabahID,
			LocalKey:   SAFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah"},
		},
		RelTargetRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         SAFieldTargetRekeningID,
			LocalKey:   SAFieldTargetRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"no_rekening"},
		},
	}
}

// Validate memvalidasi invariant domain shu_anggota sebelum write.
func (d *SHUAnggotaDescriptor) Validate(data map[string]any) error {
	if err := validateSARequired(data); err != nil {
		return err
	}
	return validateSAEnums(data)
}

func validateSARequired(data map[string]any) error {
	spID, _ := data[SAFieldSHUPeriodeID].(string)
	if spID == "" {
		return fmt.Errorf("shu_periode_id wajib diisi")
	}

	nID, _ := data[SAFieldNasabahID].(string)
	if nID == "" {
		return fmt.Errorf("nasabah_id wajib diisi")
	}

	as, ok := data[SAFieldAvgSimpanan].(float64)
	if !ok {
		return fmt.Errorf("avg_simpanan wajib diisi dan harus berupa angka")
	}
	if as < 0 {
		return fmt.Errorf("avg_simpanan tidak boleh negatif")
	}

	tt, ok := data[SAFieldTotalTransaksi].(float64)
	if !ok {
		return fmt.Errorf("total_transaksi wajib diisi dan harus berupa angka")
	}
	if tt < 0 {
		return fmt.Errorf("total_transaksi tidak boleh negatif")
	}

	ad, ok := data[SAFieldActiveDays].(float64)
	if !ok {
		return fmt.Errorf("active_days wajib diisi dan harus berupa angka")
	}
	if ad < 0 {
		return fmt.Errorf("active_days tidak boleh negatif")
	}

	jm, ok := data[SAFieldJasaModal].(float64)
	if !ok {
		return fmt.Errorf("jasa_modal wajib diisi dan harus berupa angka")
	}
	if jm < 0 {
		return fmt.Errorf("jasa_modal tidak boleh negatif")
	}

	ju, ok := data[SAFieldJasaUsaha].(float64)
	if !ok {
		return fmt.Errorf("jasa_usaha wajib diisi dan harus berupa angka")
	}
	if ju < 0 {
		return fmt.Errorf("jasa_usaha tidak boleh negatif")
	}

	ts, ok := data[SAFieldTotalSHU].(float64)
	if !ok {
		return fmt.Errorf("total_shu wajib diisi dan harus berupa angka")
	}
	if ts < 0 {
		return fmt.Errorf("total_shu tidak boleh negatif")
	}
	return nil
}

func validateSAEnums(data map[string]any) error {
	dm, _ := data[SAFieldDistMethod].(string)
	if dm != "" && !validDistMethods[dm] {
		return fmt.Errorf("distribution_method tidak valid: %q", dm)
	}
	return nil
}
