package dissolution

import "errors"

var (
	ErrNotFound                = errors.New("proses pembubaran tidak ditemukan")
	ErrInvalidDissolutionType  = errors.New("jenis pembubaran tidak valid")
	ErrInvalidStageTransition  = errors.New("transisi tahapan tidak valid")
	ErrCannotCancelCompleted   = errors.New("proses yang sudah selesai tidak dapat dibatalkan")
	ErrCannotCancelInProgress  = errors.New("proses yang sedang berjalan tidak dapat dibatalkan")
	ErrRatMeetingRequired      = errors.New("rap_meeting_id wajib untuk pembubaran sukarela (voluntary)")
	ErrReasonRequired          = errors.New("alasan pembubaran wajib diisi")
	ErrEffectiveDateRequired   = errors.New("tanggal efektif wajib diisi")
	ErrInitiatedByRequired     = errors.New("yang menginisiasi wajib diisi")
	ErrDissolutionTypeRequired = errors.New("jenis pembubaran wajib diisi")
	ErrAlreadyClosed           = errors.New("proses pembubaran sudah ditutup")
	ErrNotInitiated            = errors.New("hanya proses dengan status initiated yang dapat dibatalkan")
	ErrNetEquityRequired       = errors.New("net equity wajib diisi untuk menyelesaikan pembubaran")
	ErrMemberCountRequired     = errors.New("jumlah anggota wajib diisi untuk menghitung distribusi")
)
