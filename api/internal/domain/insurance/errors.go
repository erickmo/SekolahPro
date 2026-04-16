package insurance

import "errors"

var (
	ErrProductNotFound     = errors.New("insurance product tidak ditemukan")
	ErrPolicyNotFound      = errors.New("insurance policy tidak ditemukan")
	ErrClaimNotFound       = errors.New("insurance claim tidak ditemukan")
	ErrNameEmpty           = errors.New("name tidak boleh kosong")
	ErrCodeEmpty           = errors.New("code tidak boleh kosong")
	ErrPolicyNoEmpty       = errors.New("policy number tidak boleh kosong")
	ErrInvalidProductType  = errors.New("tipe produk tidak valid")
	ErrInvalidClaimType    = errors.New("tipe klaim tidak valid")
	ErrPolicyNotActive     = errors.New("polis tidak aktif")
	ErrClaimAlreadyPaid    = errors.New("klaim sudah dibayar")
)
