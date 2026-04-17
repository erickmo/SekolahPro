// Package jaminan adalah domain Vernon untuk jaminan/agunan koperasi.
//
// Jaminan mencatat collateral yang diserahkan nasabah untuk pengajuan pinjaman.
// Tipe: internal_balance, surat_berharga, barang_bergerak, personal_guarantee.
// Lifecycle: registered → pledged → released / foreclosed.
//
// Terdiri dari 4 tabel:
//   - jaminan         : master collateral
//   - jaminan_pinjaman: pivot jaminan ↔ pinjaman
//   - jaminan_dokumen : dokumen pendukung
//   - jaminan_valuasi : penilaian/valuasi
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package jaminan

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ──────────────────────────────────────────────────────────────────────────────
// Field name constants — jaminan
// ──────────────────────────────────────────────────────────────────────────────

const (
	FieldNasabahID      = "nasabah_id"
	FieldRekeningID     = "rekening_id"
	FieldCollateralType = "collateral_type"
	FieldDescription    = "description"
	FieldAppraisedValue = "appraised_value"
	FieldAcceptanceRate = "acceptance_rate"
	FieldCollateralVal  = "collateral_value"
	FieldStatus         = "status"
)

// Collateral type constants.
const (
	CollateralInternalBalance = "internal_balance"
	CollateralSuratBerharga   = "surat_berharga"
	CollateralBarangBergerak  = "barang_bergerak"
	CollateralPersonalGuarantee = "personal_guarantee"
)

// Status constants — jaminan.
const (
	StatusRegistered = "registered"
	StatusPledged    = "pledged"
	StatusReleased   = "released"
	StatusForeclosed = "foreclosed"
)

// Relation name constants — jaminan.
const (
	RelNasabah  = "nasabah"
	RelRekening = "rekening"
)

// validCollateralTypes berisi semua collateral_type yang valid.
var validCollateralTypes = map[string]bool{
	CollateralInternalBalance: true, CollateralSuratBerharga: true,
	CollateralBarangBergerak: true, CollateralPersonalGuarantee: true,
}

// validJaminanStatuses berisi semua status jaminan yang valid.
var validJaminanStatuses = map[string]bool{
	StatusRegistered: true, StatusPledged: true,
	StatusReleased: true, StatusForeclosed: true,
}

// ──────────────────────────────────────────────────────────────────────────────
// Field name constants — jaminan_pinjaman
// ──────────────────────────────────────────────────────────────────────────────

const (
	JPFieldJaminanID    = "jaminan_id"
	JPFieldPinjamanID   = "pinjaman_id"
	JPFieldPledgedValue = "pledged_value"
	JPFieldStatus       = "status"
)

// Status constants — jaminan_pinjaman.
const (
	JPStatusActive  = "active"
	JPStatusReleased = "released"
)

// Relation name constants — jaminan_pinjaman.
const (
	JPRelJaminan  = "jaminan"
	JPRelPinjaman = "pinjaman"
)

// validJPStatuses berisi semua status jaminan_pinjaman yang valid.
var validJPStatuses = map[string]bool{
	JPStatusActive: true, JPStatusReleased: true,
}

// ──────────────────────────────────────────────────────────────────────────────
// Field name constants — jaminan_dokumen
// ──────────────────────────────────────────────────────────────────────────────

const (
	JDFieldJaminanID   = "jaminan_id"
	JDFieldDocType     = "document_type"
	JDFieldFileURL     = "file_url"
	JDFieldFileName    = "file_name"
	JDFieldFileSize    = "file_size_bytes"
	JDFieldMimeType    = "mime_type"
)

// Document type constants.
const (
	DocFotoBarang        = "foto_barang"
	DocFotokopiBPKB      = "fotokopi_bpkb"
	DocSertifikatTanah   = "sertifikat_tanah"
	DocSuratKuasa        = "surat_kuasa"
	DocBeritaAcara       = "berita_acara"
	DocFotoIdentitas     = "foto_identitas"
	DocSuratPernyataan   = "surat_pernyataan"
	DocIjazah            = "ijazah"
	DocLainnya           = "lainnya"
)

// Relation name constants — jaminan_dokumen.
const (
	JDRelJaminan = "jaminan"
)

// validDocTypes berisi semua document_type yang valid.
var validDocTypes = map[string]bool{
	DocFotoBarang: true, DocFotokopiBPKB: true,
	DocSertifikatTanah: true, DocSuratKuasa: true,
	DocBeritaAcara: true, DocFotoIdentitas: true,
	DocSuratPernyataan: true, DocIjazah: true,
	DocLainnya: true,
}

// ──────────────────────────────────────────────────────────────────────────────
// Field name constants — jaminan_valuasi
// ──────────────────────────────────────────────────────────────────────────────

const (
	JVFieldJaminanID     = "jaminan_id"
	JVFieldAppraisedVal  = "appraised_value"
	JVFieldAcceptanceRate = "acceptance_rate"
	JVFieldCollateralVal  = "collateral_value"
	JVFieldValuationType  = "valuation_type"
)

// Valuation type constants.
const (
	ValuationInitial   = "initial"
	ValuationPeriodic  = "periodic"
	ValuationOnDemand  = "on_demand"
)

// Relation name constants — jaminan_valuasi.
const (
	JVRelJaminan = "jaminan"
)

// validValuationTypes berisi semua valuation_type yang valid.
var validValuationTypes = map[string]bool{
	ValuationInitial: true, ValuationPeriodic: true, ValuationOnDemand: true,
}

// ──────────────────────────────────────────────────────────────────────────────
// Descriptor: jaminan
// ──────────────────────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk jaminan.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "jaminan" }

// DefaultRels mendefinisikan relasi domain jaminan.
// BelongsTo nasabah (autoload) dan rekening (autoload, nullable FK).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah", "type"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         FieldRekeningID,
			LocalKey:   FieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"no_rekening", "category", "balance"},
		},
	}
}

// Validate memvalidasi invariant domain jaminan sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateJaminanRequired(data); err != nil {
		return err
	}
	if err := validateJaminanEnums(data); err != nil {
		return err
	}
	return validateJaminanAmounts(data)
}

// validateJaminanRequired memeriksa field wajib jaminan.
func validateJaminanRequired(data map[string]any) error {
	nasabahID, _ := data[FieldNasabahID].(string)
	if nasabahID == "" {
		return errors.New("nasabah_id wajib diisi")
	}

	collateralType, _ := data[FieldCollateralType].(string)
	if collateralType == "" {
		return errors.New("collateral_type wajib diisi")
	}

	desc, _ := data[FieldDescription].(string)
	if desc == "" {
		return errors.New("description wajib diisi")
	}
	return nil
}

// validateJaminanEnums memeriksa enum fields jaminan.
func validateJaminanEnums(data map[string]any) error {
	collateralType, _ := data[FieldCollateralType].(string)
	if collateralType != "" && !validCollateralTypes[collateralType] {
		return fmt.Errorf("collateral_type tidak valid: %q", collateralType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validJaminanStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus registered/pledged/released/foreclosed)", status)
	}
	return nil
}

// validateJaminanAmounts memeriksa amount dan rate fields jaminan.
func validateJaminanAmounts(data map[string]any) error {
	appraised, ok := data[FieldAppraisedValue].(float64)
	if ok && appraised < 0 {
		return errors.New("appraised_value tidak boleh negatif")
	}

	rate, ok := data[FieldAcceptanceRate].(float64)
	if ok && (rate < 0 || rate > 1) {
		return errors.New("acceptance_rate harus antara 0 dan 1")
	}

	collateralVal, ok := data[FieldCollateralVal].(float64)
	if ok && collateralVal < 0 {
		return errors.New("collateral_value tidak boleh negatif")
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Descriptor: jaminan_pinjaman
// ──────────────────────────────────────────────────────────────────────────────

// PinjamanDescriptor mengimplementasi vernon.DomainDescriptor untuk jaminan_pinjaman.
type PinjamanDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PinjamanDescriptor) TableName() string { return "jaminan_pinjaman" }

// DefaultRels mendefinisikan relasi jaminan_pinjaman.
// BelongsTo jaminan (autoload) dan pinjaman (autoload).
func (d *PinjamanDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		JPRelJaminan: {
			Domain:     "jaminan",
			Type:       vernon.RelBelongsTo,
			FK:         JPFieldJaminanID,
			LocalKey:   JPFieldJaminanID,
			ForeignKey: "id",
			IsAutoload: true,
		},
		JPRelPinjaman: {
			Domain:     "pinjaman",
			Type:       vernon.RelBelongsTo,
			FK:         JPFieldPinjamanID,
			LocalKey:   JPFieldPinjamanID,
			ForeignKey: "id",
			IsAutoload: true,
		},
	}
}

// Validate memvalidasi invariant domain jaminan_pinjaman.
func (d *PinjamanDescriptor) Validate(data map[string]any) error {
	if err := validateJPRequired(data); err != nil {
		return err
	}
	return validateJPEnums(data)
}

// validateJPRequired memeriksa field wajib jaminan_pinjaman.
func validateJPRequired(data map[string]any) error {
	jaminanID, _ := data[JPFieldJaminanID].(string)
	if jaminanID == "" {
		return errors.New("jaminan_id wajib diisi")
	}

	pinjamanID, _ := data[JPFieldPinjamanID].(string)
	if pinjamanID == "" {
		return errors.New("pinjaman_id wajib diisi")
	}
	return nil
}

// validateJPEnums memeriksa enum fields jaminan_pinjaman.
func validateJPEnums(data map[string]any) error {
	pledgedVal, ok := data[JPFieldPledgedValue].(float64)
	if ok && pledgedVal < 0 {
		return errors.New("pledged_value tidak boleh negatif")
	}

	status, _ := data[JPFieldStatus].(string)
	if status != "" && !validJPStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus active/released)", status)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Descriptor: jaminan_dokumen
// ──────────────────────────────────────────────────────────────────────────────

// DokumenDescriptor mengimplementasi vernon.DomainDescriptor untuk jaminan_dokumen.
type DokumenDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *DokumenDescriptor) TableName() string { return "jaminan_dokumen" }

// DefaultRels mendefinisikan relasi jaminan_dokumen.
// BelongsTo jaminan (autoload).
func (d *DokumenDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		JDRelJaminan: {
			Domain:     "jaminan",
			Type:       vernon.RelBelongsTo,
			FK:         JDFieldJaminanID,
			LocalKey:   JDFieldJaminanID,
			ForeignKey: "id",
			IsAutoload: true,
		},
	}
}

// Validate memvalidasi invariant domain jaminan_dokumen.
func (d *DokumenDescriptor) Validate(data map[string]any) error {
	if err := validateJDRequired(data); err != nil {
		return err
	}
	return validateJDFile(data)
}

// validateJDRequired memeriksa field wajib jaminan_dokumen.
func validateJDRequired(data map[string]any) error {
	jaminanID, _ := data[JDFieldJaminanID].(string)
	if jaminanID == "" {
		return errors.New("jaminan_id wajib diisi")
	}

	docType, _ := data[JDFieldDocType].(string)
	if docType == "" {
		return errors.New("document_type wajib diisi")
	}
	if !validDocTypes[docType] {
		return fmt.Errorf("document_type tidak valid: %q", docType)
	}

	fileURL, _ := data[JDFieldFileURL].(string)
	if fileURL == "" {
		return errors.New("file_url wajib diisi")
	}

	fileName, _ := data[JDFieldFileName].(string)
	if fileName == "" {
		return errors.New("file_name wajib diisi")
	}

	mimeType, _ := data[JDFieldMimeType].(string)
	if mimeType == "" {
		return errors.New("mime_type wajib diisi")
	}
	return nil
}

// validateJDFile memeriksa file_size_bytes > 0.
func validateJDFile(data map[string]any) error {
	fileSize, ok := data[JDFieldFileSize].(float64)
	if ok && fileSize <= 0 {
		return errors.New("file_size_bytes harus lebih dari 0")
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Descriptor: jaminan_valuasi
// ──────────────────────────────────────────────────────────────────────────────

// ValuasiDescriptor mengimplementasi vernon.DomainDescriptor untuk jaminan_valuasi.
type ValuasiDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ValuasiDescriptor) TableName() string { return "jaminan_valuasi" }

// DefaultRels mendefinisikan relasi jaminan_valuasi.
// BelongsTo jaminan (autoload).
func (d *ValuasiDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		JVRelJaminan: {
			Domain:     "jaminan",
			Type:       vernon.RelBelongsTo,
			FK:         JVFieldJaminanID,
			LocalKey:   JVFieldJaminanID,
			ForeignKey: "id",
			IsAutoload: true,
		},
	}
}

// Validate memvalidasi invariant domain jaminan_valuasi.
func (d *ValuasiDescriptor) Validate(data map[string]any) error {
	if err := validateJVRequired(data); err != nil {
		return err
	}
	if err := validateJVAmounts(data); err != nil {
		return err
	}
	return validateJVEnums(data)
}

// validateJVRequired memeriksa field wajib jaminan_valuasi.
func validateJVRequired(data map[string]any) error {
	jaminanID, _ := data[JVFieldJaminanID].(string)
	if jaminanID == "" {
		return errors.New("jaminan_id wajib diisi")
	}

	valuationType, _ := data[JVFieldValuationType].(string)
	if valuationType == "" {
		return errors.New("valuation_type wajib diisi")
	}
	return nil
}

// validateJVAmounts memeriksa amount dan rate fields jaminan_valuasi.
func validateJVAmounts(data map[string]any) error {
	appraised, ok := data[JVFieldAppraisedVal].(float64)
	if ok && appraised < 0 {
		return errors.New("appraised_value tidak boleh negatif")
	}

	rate, ok := data[JVFieldAcceptanceRate].(float64)
	if ok && (rate < 0 || rate > 1) {
		return errors.New("acceptance_rate harus antara 0 dan 1")
	}

	collateralVal, ok := data[JVFieldCollateralVal].(float64)
	if ok && collateralVal < 0 {
		return errors.New("collateral_value tidak boleh negatif")
	}
	return nil
}

// validateJVEnums memeriksa valuation_type enum.
func validateJVEnums(data map[string]any) error {
	valuationType, _ := data[JVFieldValuationType].(string)
	if valuationType != "" && !validValuationTypes[valuationType] {
		return fmt.Errorf("valuation_type tidak valid: %q (harus initial/periodic/on_demand)", valuationType)
	}
	return nil
}
