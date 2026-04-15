package role

import "errors"

var (
	ErrNotFound          = errors.New("role tidak ditemukan")
	ErrNameEmpty         = errors.New("name wajib diisi")
	ErrCodeEmpty         = errors.New("code wajib diisi")
	ErrCodeExists        = errors.New("code sudah digunakan di company ini")
	ErrInvalidRoleType   = errors.New("role_type tidak valid")
	ErrSystemRoleDelete  = errors.New("system role tidak bisa dihapus")
	ErrSystemRoleModify  = errors.New("system role tidak bisa dimodifikasi")
	ErrAlreadyAssigned   = errors.New("user sudah memiliki role ini di company ini")
	ErrUserRoleNotFound  = errors.New("user role assignment tidak ditemukan")
)
