package laboratory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/laboratory"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── LabDescriptor metadata tests ──────────────────────────────────────────────

func TestLabDescriptor_TableName(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	assert.Equal(t, "laboratories", d.TableName())
}

func TestLabDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laboratory.LabDescriptor{}
}

func TestLabDescriptor_DefaultRels_Count(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "laboratories should have no relations")
}

// ── LabDescriptor validation tests ────────────────────────────────────────────

func validLabData() map[string]any {
	return map[string]any{
		laboratory.LabFieldName: "Lab IPA 1",
		laboratory.LabFieldType: laboratory.LabTypeIPA,
	}
}

func TestLabDescriptor_Validate_Success(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	assert.NoError(t, d.Validate(validLabData()))
}

func TestLabDescriptor_Validate_AllLabTypes(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	types := []string{
		laboratory.LabTypeIPA, laboratory.LabTypeKomputer,
		laboratory.LabTypeBahasa, laboratory.LabTypeMultimedia,
	}
	for _, lt := range types {
		data := validLabData()
		data[laboratory.LabFieldType] = lt
		err := d.Validate(data)
		assert.NoError(t, err, "type=%q should be valid", lt)
	}
}

func TestLabDescriptor_Validate_MissingName(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	data := validLabData()
	delete(data, laboratory.LabFieldName)
	assert.ErrorContains(t, d.Validate(data), "name wajib diisi")
}

func TestLabDescriptor_Validate_MissingType(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	data := validLabData()
	delete(data, laboratory.LabFieldType)
	assert.ErrorContains(t, d.Validate(data), "type wajib diisi")
}

func TestLabDescriptor_Validate_InvalidType(t *testing.T) {
	d := &laboratory.LabDescriptor{}
	data := validLabData()
	data[laboratory.LabFieldType] = "art"
	assert.ErrorContains(t, d.Validate(data), "type tidak valid")
}

// ── EquipmentDescriptor metadata tests ────────────────────────────────────────

func TestEquipmentDescriptor_TableName(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	assert.Equal(t, "lab_equipment", d.TableName())
}

func TestEquipmentDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laboratory.EquipmentDescriptor{}
}

func TestEquipmentDescriptor_DefaultRels_Count(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "lab_equipment should have 1 BelongsTo relation")
}

func TestEquipmentDescriptor_DefaultRels_LabRelation(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[laboratory.RelLab]
	require.True(t, ok, "should have lab relation")

	assert.Equal(t, "laboratories", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, laboratory.EqFieldLabID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "type", "room_number"}, rel.Fields)
}

// ── EquipmentDescriptor validation tests ──────────────────────────────────────

func validEquipmentData() map[string]any {
	return map[string]any{
		laboratory.EqFieldLabID:     "00000000-0000-0000-0000-000000000001",
		laboratory.EqFieldName:      "Mikroskop Olympus",
		laboratory.EqFieldCategory:  laboratory.EqCatOptical,
	}
}

func TestEquipmentDescriptor_Validate_Success(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	assert.NoError(t, d.Validate(validEquipmentData()))
}

func TestEquipmentDescriptor_Validate_MissingLabID(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	data := validEquipmentData()
	delete(data, laboratory.EqFieldLabID)
	assert.ErrorContains(t, d.Validate(data), "lab_id wajib diisi")
}

func TestEquipmentDescriptor_Validate_MissingName(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	data := validEquipmentData()
	delete(data, laboratory.EqFieldName)
	assert.ErrorContains(t, d.Validate(data), "name wajib diisi")
}

func TestEquipmentDescriptor_Validate_MissingCategory(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	data := validEquipmentData()
	delete(data, laboratory.EqFieldCategory)
	assert.ErrorContains(t, d.Validate(data), "category wajib diisi")
}

func TestEquipmentDescriptor_Validate_InvalidCategory(t *testing.T) {
	d := &laboratory.EquipmentDescriptor{}
	data := validEquipmentData()
	data[laboratory.EqFieldCategory] = "robotic"
	assert.ErrorContains(t, d.Validate(data), "category tidak valid")
}

// ── UsageLogDescriptor metadata tests ─────────────────────────────────────────

func TestUsageLogDescriptor_TableName(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	assert.Equal(t, "lab_usage_logs", d.TableName())
}

func TestUsageLogDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laboratory.UsageLogDescriptor{}
}

func TestUsageLogDescriptor_DefaultRels_Count(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "lab_usage_logs should have 4 BelongsTo relations")
}

func TestUsageLogDescriptor_DefaultRels_LabRelation(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[laboratory.RelLab]
	require.True(t, ok, "should have lab relation")

	assert.Equal(t, "laboratories", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, laboratory.ULFieldLabID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "type"}, rel.Fields)
}

// ── UsageLogDescriptor validation tests ───────────────────────────────────────

func validUsageLogData() map[string]any {
	return map[string]any{
		laboratory.ULFieldLabID:     "00000000-0000-0000-0000-000000000001",
		laboratory.ULFieldTeacherID: "00000000-0000-0000-0000-000000000002",
		laboratory.ULFieldUsageDate: "2026-04-01",
	}
}

func TestUsageLogDescriptor_Validate_Success(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	assert.NoError(t, d.Validate(validUsageLogData()))
}

func TestUsageLogDescriptor_Validate_MissingLabID(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	data := validUsageLogData()
	delete(data, laboratory.ULFieldLabID)
	assert.ErrorContains(t, d.Validate(data), "lab_id wajib diisi")
}

func TestUsageLogDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	data := validUsageLogData()
	delete(data, laboratory.ULFieldTeacherID)
	assert.ErrorContains(t, d.Validate(data), "teacher_id wajib diisi")
}

func TestUsageLogDescriptor_Validate_MissingUsageDate(t *testing.T) {
	d := &laboratory.UsageLogDescriptor{}
	data := validUsageLogData()
	delete(data, laboratory.ULFieldUsageDate)
	assert.ErrorContains(t, d.Validate(data), "usage_date wajib diisi")
}
