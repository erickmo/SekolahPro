package biometric

import "github.com/google/uuid"

// BiometricEnrolledEvent dipublikasikan saat biometric enrollment baru berhasil dibuat.
type BiometricEnrolledEvent struct {
	TenantID      uuid.UUID     `json:"tenant_id"`
	CompanyID     uuid.UUID     `json:"company_id"`
	ID            uuid.UUID     `json:"id"`
	NasabahID     uuid.UUID     `json:"nasabah_id"`
	BiometricType BiometricType `json:"biometric_type"`
}

func (e BiometricEnrolledEvent) EventName() string   { return "biometric.enrolled" }
func (e BiometricEnrolledEvent) AggregateID() string { return e.ID.String() }

// BiometricVerifiedEvent dipublikasikan saat verifikasi biometric berhasil.
type BiometricVerifiedEvent struct {
	TenantID      uuid.UUID     `json:"tenant_id"`
	CompanyID     uuid.UUID     `json:"company_id"`
	ID            uuid.UUID     `json:"id"`
	NasabahID     uuid.UUID     `json:"nasabah_id"`
	BiometricType BiometricType `json:"biometric_type"`
}

func (e BiometricVerifiedEvent) EventName() string   { return "biometric.verified" }
func (e BiometricVerifiedEvent) AggregateID() string { return e.ID.String() }

// BiometricDeactivatedEvent dipublikasikan saat biometric enrollment dinonaktifkan.
type BiometricDeactivatedEvent struct {
	TenantID      uuid.UUID     `json:"tenant_id"`
	CompanyID     uuid.UUID     `json:"company_id"`
	ID            uuid.UUID     `json:"id"`
	NasabahID     uuid.UUID     `json:"nasabah_id"`
	BiometricType BiometricType `json:"biometric_type"`
}

func (e BiometricDeactivatedEvent) EventName() string   { return "biometric.deactivated" }
func (e BiometricDeactivatedEvent) AggregateID() string { return e.ID.String() }
