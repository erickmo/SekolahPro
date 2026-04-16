package enrollment

import "github.com/google/uuid"

// EnrollmentCreatedEvent dipublikasikan saat EducationEnrollment baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type EnrollmentCreatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	NasabahID uuid.UUID `json:"nasabah_id"`
}

func (e EnrollmentCreatedEvent) EventName() string   { return "education_enrollment.created" }
func (e EnrollmentCreatedEvent) AggregateID() string { return e.ID.String() }

// EnrollmentCompletedEvent dipublikasikan saat EducationEnrollment berhasil diselesaikan.
type EnrollmentCompletedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	NasabahID uuid.UUID `json:"nasabah_id"`
	Score     float64   `json:"score"`
}

func (e EnrollmentCompletedEvent) EventName() string   { return "education_enrollment.completed" }
func (e EnrollmentCompletedEvent) AggregateID() string { return e.ID.String() }
