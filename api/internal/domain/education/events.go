package education

import "github.com/google/uuid"

// CourseCreatedEvent dipublikasikan saat EducationCourse baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type CourseCreatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
}

func (e CourseCreatedEvent) EventName() string   { return "education_course.created" }
func (e CourseCreatedEvent) AggregateID() string { return e.ID.String() }

// CourseUpdatedEvent dipublikasikan saat EducationCourse berhasil diupdate.
type CourseUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
}

func (e CourseUpdatedEvent) EventName() string   { return "education_course.updated" }
func (e CourseUpdatedEvent) AggregateID() string { return e.ID.String() }
