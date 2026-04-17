package achievement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/achievement"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &achievement.Descriptor{}
	assert.Equal(t, "achievements", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &achievement.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &achievement.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "achievement should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &achievement.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[achievement.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, achievement.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validAchievementData() map[string]any {
	return map[string]any{
		achievement.FieldStudentID:      "00000000-0000-0000-0000-000000000001",
		achievement.FieldTitle:          "Juara 1 Olimpiade Matematika",
		achievement.FieldAchievementType: achievement.AchievementTypeAcademic,
		achievement.FieldLevel:          achievement.LevelNational,
		achievement.FieldDate:           "2026-03-10",
		achievement.FieldOrganizer:      "Kemendikbud",
		achievement.FieldDescription:    "Olimpiade Sains Nasional",
		achievement.FieldCertificateURL: "https://example.com/cert.pdf",
		achievement.FieldStatus:         "verified",
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &achievement.Descriptor{}
	err := d.Validate(validAchievementData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllAchievementTypes(t *testing.T) {
	d := &achievement.Descriptor{}
	types := []string{
		achievement.AchievementTypeAcademic, achievement.AchievementTypeSport,
		achievement.AchievementTypeArt, achievement.AchievementTypeTechnology,
		achievement.AchievementTypeSocial,
	}
	for _, at := range types {
		data := validAchievementData()
		data[achievement.FieldAchievementType] = at
		err := d.Validate(data)
		assert.NoError(t, err, "achievement_type=%q should be valid", at)
	}
}

func TestDescriptor_Validate_AllLevels(t *testing.T) {
	d := &achievement.Descriptor{}
	levels := []string{
		achievement.LevelSchool, achievement.LevelDistrict,
		achievement.LevelProvince, achievement.LevelNational,
		achievement.LevelInternational,
	}
	for _, lvl := range levels {
		data := validAchievementData()
		data[achievement.FieldLevel] = lvl
		err := d.Validate(data)
		assert.NoError(t, err, "level=%q should be valid", lvl)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &achievement.Descriptor{}
	data := validAchievementData()
	delete(data, achievement.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_InvalidAchievementType(t *testing.T) {
	d := &achievement.Descriptor{}
	data := validAchievementData()
	data[achievement.FieldAchievementType] = "gaming"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "achievement_type tidak valid")
}

func TestDescriptor_Validate_InvalidLevel(t *testing.T) {
	d := &achievement.Descriptor{}
	data := validAchievementData()
	data[achievement.FieldLevel] = "global"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "level tidak valid")
}

func TestDescriptor_Validate_EmptyAchievementType_OK(t *testing.T) {
	d := &achievement.Descriptor{}
	data := validAchievementData()
	data[achievement.FieldAchievementType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty achievement_type should be valid (optional)")
}

func TestDescriptor_Validate_EmptyLevel_OK(t *testing.T) {
	d := &achievement.Descriptor{}
	data := validAchievementData()
	data[achievement.FieldLevel] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty level should be valid (optional)")
}
