// Package achievement adalah domain Vernon untuk prestasi siswa.
//
// Achievement memiliki 1 BelongsTo autoload: student.
// Digunakan untuk tracking prestasi akademik dan non-akademik siswa.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package achievement

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID     = "student_id"
	FieldTitle         = "title"
	FieldAchievementType = "achievement_type"
	FieldLevel         = "level"
	FieldDate          = "date"
	FieldOrganizer     = "organizer"
	FieldDescription   = "description"
	FieldCertificateURL = "certificate_url"
	FieldStatus        = "status"
)

// Achievement type constants.
const (
	AchievementTypeAcademic   = "academic"
	AchievementTypeSport      = "sport"
	AchievementTypeArt        = "art"
	AchievementTypeTechnology = "technology"
	AchievementTypeSocial     = "social"
)

// Level constants.
const (
	LevelSchool    = "school"
	LevelDistrict  = "district"
	LevelProvince  = "province"
	LevelNational  = "national"
	LevelInternational = "international"
)

// Relation name constants.
const (
	RelStudent = "student"
)

var validAchievementTypes = map[string]bool{
	AchievementTypeAcademic: true, AchievementTypeSport: true,
	AchievementTypeArt: true, AchievementTypeTechnology: true,
	AchievementTypeSocial: true,
}

var validLevels = map[string]bool{
	LevelSchool: true, LevelDistrict: true,
	LevelProvince: true, LevelNational: true, LevelInternational: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk achievements.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "achievements" }

// DefaultRels mendefinisikan relasi domain ini.
// 1 BelongsTo autoload: student.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	achievementType, _ := data[FieldAchievementType].(string)
	if achievementType != "" && !validAchievementTypes[achievementType] {
		return fmt.Errorf("achievement_type tidak valid: %q", achievementType)
	}

	level, _ := data[FieldLevel].(string)
	if level != "" && !validLevels[level] {
		return fmt.Errorf("level tidak valid: %q", level)
	}
	return nil
}
