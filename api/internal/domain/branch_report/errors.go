package branch_report

import "errors"

var (
	ErrNotFound                      = errors.New("laporan keuangan tidak ditemukan")
	ErrDuplicatePeriod               = errors.New("laporan untuk periode ini sudah ada")
	ErrInvalidPeriodRange            = errors.New("periode awal harus sebelum periode akhir")
	ErrInvalidPeriodType             = errors.New("jenis periode tidak valid")
	ErrAlreadyApproved               = errors.New("laporan sudah disetujui")
	ErrCannotApproveConsolidated     = errors.New("laporan konsolidasi tidak dapat disetujui langsung")
	ErrRuleNotFound                  = errors.New("elimination rule tidak ditemukan")
	ErrInvalidRuleType               = errors.New("jenis elimination rule tidak valid")
	ErrRuleNameEmpty                 = errors.New("nama rule tidak boleh kosong")
	ErrSameBranchTransfer            = errors.New("cabang asal dan tujuan tidak boleh sama")
	ErrNoBranchesForConsolidation    = errors.New("tidak ada data cabang untuk konsolidasi")
)
