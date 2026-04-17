package teller_session_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/boilerplate/internal/domain/teller_session"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &teller_session.Descriptor{}
	assert.Equal(t, "teller_sessions", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teller_session.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teller_session.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "teller_session is a root entity and should have 0 relations")
}

// ── Validation tests ────────────────────────────────────────────────────────

func validTellerSessionData() map[string]any {
	return map[string]any{
		teller_session.FieldUserID:      "00000000-0000-0000-0000-000000000001",
		teller_session.FieldSessionDate: "2026-04-17",
		teller_session.FieldStatus:      teller_session.StatusOpen,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &teller_session.Descriptor{}
	err := d.Validate(validTellerSessionData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingUserID(t *testing.T) {
	d := &teller_session.Descriptor{}
	data := validTellerSessionData()
	delete(data, teller_session.FieldUserID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "user_id wajib diisi")
}

func TestDescriptor_Validate_MissingSessionDate(t *testing.T) {
	d := &teller_session.Descriptor{}
	data := validTellerSessionData()
	delete(data, teller_session.FieldSessionDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "session_date wajib diisi")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &teller_session.Descriptor{}
	data := validTellerSessionData()
	data[teller_session.FieldStatus] = "suspended"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &teller_session.Descriptor{}
	statuses := []string{teller_session.StatusOpen, teller_session.StatusClosed, teller_session.StatusReconciled}
	for _, status := range statuses {
		data := validTellerSessionData()
		data[teller_session.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_EmptyStatus_OK(t *testing.T) {
	d := &teller_session.Descriptor{}
	data := validTellerSessionData()
	data[teller_session.FieldStatus] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty status should be valid (optional)")
}
