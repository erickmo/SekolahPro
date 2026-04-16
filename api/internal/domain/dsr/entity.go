// Package dsr mendefinisikan domain untuk Data Subject Request (permintaan data subjek).
//
// Implementasi UU PDP No. 27/2022: hak akses, koreksi, penghapusan, portabilitas data.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package dsr

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// RequestType mendefinisikan jenis permintaan data subjek sesuai UU PDP.
type RequestType string

const (
	RequestTypeAccess      RequestType = "access"
	RequestTypeCorrection  RequestType = "correction"
	RequestTypeDeletion    RequestType = "deletion"
	RequestTypePortability RequestType = "portability"
	RequestTypeObjection   RequestType = "objection"
	RequestTypeRestriction RequestType = "restriction"
)

// RequestStatus mendefinisikan status permintaan.
type RequestStatus string

const (
	RequestStatusPending      RequestStatus = "pending"
	RequestStatusVerified     RequestStatus = "verified"
	RequestStatusInProgress   RequestStatus = "in_progress"
	RequestStatusCompleted    RequestStatus = "completed"
	RequestStatusRejected     RequestStatus = "rejected"
	RequestStatusCancelled    RequestStatus = "cancelled"
)

// DataSubjectRequest adalah entity utama untuk permintaan data subjek.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type DataSubjectRequest struct {
	ID              uuid.UUID     `db:"id"`
	RequestType     RequestType   `db:"request_type"`
	Status          RequestStatus `db:"status"`
	RequestorName   string        `db:"requestor_name"`
	RequestorEmail  string        `db:"requestor_email"`
	SubjectID       uuid.UUID     `db:"subject_id"`
	SubjectType     string        `db:"subject_type"`
	Description     string        `db:"description"`
	VerifiedAt      *time.Time    `db:"verified_at"`
	VerifiedBy      *uuid.UUID    `db:"verified_by"`
	CompletedAt     *time.Time    `db:"completed_at"`
	CompletedBy     *uuid.UUID    `db:"completed_by"`
	RejectionReason string        `db:"rejection_reason"`
	ResponseData    string        `db:"response_data"`
	DueDate         time.Time     `db:"due_date"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
	DeletedAt       *time.Time    `db:"deleted_at"`
}

// DSRFilter berisi parameter filter opsional untuk list query.
type DSRFilter struct {
	RequestType *RequestType
	Status      *RequestStatus
	SubjectID   *uuid.UUID
}

// WriteRepository mendefinisikan operasi write untuk domain DSR.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *DataSubjectRequest) error
	Update(ctx context.Context, s scope.Scope, e *DataSubjectRequest) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain DSR.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*DataSubjectRequest, error)
	List(ctx context.Context, s scope.Scope, filter DSRFilter, limit, offset int, sortBy, order string) ([]*DataSubjectRequest, int, error)
}
