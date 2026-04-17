package health_record_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/health_record"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &health_record.Descriptor{}
	assert.Equal(t, "health_records", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &health_record.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &health_record.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "health_record should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &health_record.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[health_record.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, health_record.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validHealthRecordData() map[string]any {
	return map[string]any{
		health_record.FieldStudentID:    "00000000-0000-0000-0000-000000000001",
		health_record.FieldRecordType:   health_record.RecordTypeGeneral,
		health_record.FieldRecordDate:   "2026-01-15",
		health_record.FieldDescription:  "Pemeriksaan rutin",
		health_record.FieldDiagnosis:    "Sehat",
		health_record.FieldTreatment:    "Tidak ada",
		health_record.FieldDoctorName:   "Dr. Siti",
		health_record.FieldFollowUpDate: "2026-06-15",
		health_record.FieldStatus:       "completed",
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &health_record.Descriptor{}
	err := d.Validate(validHealthRecordData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllRecordTypes(t *testing.T) {
	d := &health_record.Descriptor{}
	recordTypes := []string{
		health_record.RecordTypeGeneral, health_record.RecordTypeVaccination,
		health_record.RecordTypeAllergy, health_record.RecordTypeInjury,
		health_record.RecordTypeChronic, health_record.RecordTypeVision,
		health_record.RecordTypeDental,
	}
	for _, rt := range recordTypes {
		data := validHealthRecordData()
		data[health_record.FieldRecordType] = rt
		err := d.Validate(data)
		assert.NoError(t, err, "record_type=%q should be valid", rt)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &health_record.Descriptor{}
	data := validHealthRecordData()
	delete(data, health_record.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_InvalidRecordType(t *testing.T) {
	d := &health_record.Descriptor{}
	data := validHealthRecordData()
	data[health_record.FieldRecordType] = "surgery"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "record_type tidak valid")
}

func TestDescriptor_Validate_EmptyRecordType_OK(t *testing.T) {
	d := &health_record.Descriptor{}
	data := validHealthRecordData()
	data[health_record.FieldRecordType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty record_type should be valid (optional)")
}
