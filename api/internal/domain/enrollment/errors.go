package enrollment

import "errors"

var (
	ErrEnrollmentNotFound = errors.New("pendaftaran kursus tidak ditemukan")
	ErrKPINotFound        = errors.New("KPI pendidikan tidak ditemukan")
	ErrInvalidStatus      = errors.New("status pendaftaran tidak valid")
	ErrInvalidScore       = errors.New("nilai harus antara 0-100")
	ErrInvalidMetricType  = errors.New("jenis metrik tidak valid")
	ErrCourseRequired     = errors.New("kursus wajib dipilih")
	ErrNasabahRequired    = errors.New("anggota wajib dipilih")
)
