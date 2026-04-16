package nasabah_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/nasabah"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &nasabah.Descriptor{}
	assert.Equal(t, "nasabah", d.TableName())
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &nasabah.Descriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "nasabah is a root entity — should have no relations")
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &nasabah.Descriptor{}
}

// ── Valid data — all enum combinations (type x gender) ──────────────────────

func TestDescriptor_Validate_AllTypeGenderCombinations(t *testing.T) {
	d := &nasabah.Descriptor{}

	types := []string{
		nasabah.TypeSiswa,
		nasabah.TypeGuru,
		nasabah.TypeStaff,
		nasabah.TypeUmum,
	}
	genders := []string{
		nasabah.GenderL,
		nasabah.GenderP,
	}

	for _, typ := range types {
		for _, gender := range genders {
			data := map[string]any{
				nasabah.FieldNIK:      "3201234567890001",
				nasabah.FieldFullName: "Test User",
				nasabah.FieldNoNasabah: "NSB-001",
				nasabah.FieldGender:   gender,
				nasabah.FieldType:     typ,
			}
			err := d.Validate(data)
			assert.NoError(t, err, "type=%q gender=%q should be valid", typ, gender)
		}
	}
}

func TestDescriptor_Validate_ValidFullData(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:             "3201234567890001",
		nasabah.FieldFullName:        "Budi Santoso",
		nasabah.FieldNoNasabah:       "NSB-001",
		nasabah.FieldType:            nasabah.TypeSiswa,
		nasabah.FieldGender:          nasabah.GenderL,
		nasabah.FieldBirthPlace:      "Bandung",
		nasabah.FieldBirthDate:       "2005-01-15",
		nasabah.FieldPhone:           "08123456789",
		nasabah.FieldEmail:           "budi@sekolah.sch.id",
		nasabah.FieldAddress:         "Jl. Merdeka No. 1",
		nasabah.FieldStatus:          nasabah.StatusActive,
		nasabah.FieldJoinDate:        "2024-01-01",
		nasabah.FieldMotherMaidenName: "Siti Aminah",
	}
	err := d.Validate(data)
	assert.NoError(t, err)
}

// ── Valid status combinations ────────────────────────────────────────────────

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &nasabah.Descriptor{}
	statuses := []string{nasabah.StatusActive, nasabah.StatusInactive, nasabah.StatusBlacklisted}

	for _, status := range statuses {
		data := map[string]any{
			nasabah.FieldNIK:      "3201234567890001",
			nasabah.FieldFullName: "Test User",
			nasabah.FieldType:     nasabah.TypeSiswa,
			nasabah.FieldGender:   nasabah.GenderL,
			nasabah.FieldStatus:   status,
		}
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_StatusOptional(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Test User",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	assert.NoError(t, err, "status is optional and can be omitted")
}

// ── Missing required fields ─────────────────────────────────────────────────

func TestDescriptor_Validate_MissingRequired_NIK(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik wajib diisi")
}

func TestDescriptor_Validate_MissingRequired_Type(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type wajib diisi")
}

func TestDescriptor_Validate_MissingRequired_Gender(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     nasabah.TypeSiswa,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gender wajib diisi")
}

// ── NIK boundary conditions ─────────────────────────────────────────────────

func TestDescriptor_Validate_NIK_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name    string
		nik     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid 16 digit NIK",
			nik:     "3201234567890001",
			wantErr: false,
		},
		{
			name:    "all zeros NIK",
			nik:     "0000000000000000",
			wantErr: false,
		},
		{
			name:    "all nines NIK",
			nik:     "9999999999999999",
			wantErr: false,
		},
		{
			name:    "15 digits — too short",
			nik:     "320123456789000",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "17 digits — too long",
			nik:     "32012345678900011",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "empty NIK",
			nik:     "",
			wantErr: true,
			errMsg:  "nik wajib diisi",
		},
		{
			name:    "letters in NIK",
			nik:     "320123456789000A",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "spaces in NIK",
			nik:     "3201 234 567 890",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "special chars in NIK",
			nik:     "32012345678900-1",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "single digit",
			nik:     "3",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
		{
			name:    "16 chars with mixed digits and letters",
			nik:     "abcdefghijklmnop",
			wantErr: true,
			errMsg:  "nik tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &nasabah.Descriptor{}
			data := map[string]any{
				nasabah.FieldNIK:      tt.nik,
				nasabah.FieldFullName: "Test User",
				nasabah.FieldType:     nasabah.TypeSiswa,
				nasabah.FieldGender:   nasabah.GenderL,
			}
			err := d.Validate(data)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── Invalid enum values ─────────────────────────────────────────────────────

func TestDescriptor_Validate_InvalidType(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     "mahasiswa",
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type tidak valid")
}

func TestDescriptor_Validate_InvalidGender(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   "X",
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gender tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   nasabah.GenderL,
		nasabah.FieldStatus:   "unknown",
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status tidak valid")
}

func TestDescriptor_Validate_GenderCaseSensitive(t *testing.T) {
	d := &nasabah.Descriptor{}
	// Lowercase "l" and "p" should be rejected.
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Budi Santoso",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   "l",
	}
	err := d.Validate(data)
	require.Error(t, err, "gender should be case-sensitive — 'l' is not valid")
	assert.Contains(t, err.Error(), "gender tidak valid")
}

// ── Empty/nil data map ──────────────────────────────────────────────────────

func TestDescriptor_Validate_NilData(t *testing.T) {
	d := &nasabah.Descriptor{}
	err := d.Validate(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik wajib diisi")
}

func TestDescriptor_Validate_EmptyData(t *testing.T) {
	d := &nasabah.Descriptor{}
	err := d.Validate(map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik wajib diisi")
}

// ── Validation order: NIK first, then type, then gender ─────────────────────

func TestDescriptor_Validate_ValidationOrder(t *testing.T) {
	d := &nasabah.Descriptor{}

	// Missing NIK should be the first error reported
	data := map[string]any{
		nasabah.FieldType:   "invalid",
		nasabah.FieldGender: "invalid",
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik wajib diisi", "NIK should be validated first")
}

func TestDescriptor_Validate_NIKValid_InvalidType(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		nasabah.FieldNIK:      "3201234567890001",
		nasabah.FieldFullName: "Test",
		nasabah.FieldType:     "invalid",
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type tidak valid")
}

// ── Empty string required fields ─────────────────────────────────────────────

func TestDescriptor_Validate_EmptyStringRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr string
	}{
		{
			name: "empty NIK",
			data: map[string]any{
				nasabah.FieldNIK:      "",
				nasabah.FieldFullName: "Test",
				nasabah.FieldType:     nasabah.TypeSiswa,
				nasabah.FieldGender:   nasabah.GenderL,
			},
			wantErr: "nik wajib diisi",
		},
		{
			name: "empty type",
			data: map[string]any{
				nasabah.FieldNIK:      "3201234567890001",
				nasabah.FieldFullName: "Test",
				nasabah.FieldType:     "",
				nasabah.FieldGender:   nasabah.GenderL,
			},
			wantErr: "type wajib diisi",
		},
		{
			name: "empty gender",
			data: map[string]any{
				nasabah.FieldNIK:      "3201234567890001",
				nasabah.FieldFullName: "Test",
				nasabah.FieldType:     nasabah.TypeSiswa,
				nasabah.FieldGender:   "",
			},
			wantErr: "gender wajib diisi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &nasabah.Descriptor{}
			err := d.Validate(tt.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// ── NIK very long string ─────────────────────────────────────────────────────

func TestDescriptor_Validate_NIK_VeryLong(t *testing.T) {
	d := &nasabah.Descriptor{}
	longNIK := strings.Repeat("1", 100)
	data := map[string]any{
		nasabah.FieldNIK:      longNIK,
		nasabah.FieldFullName: "Test User",
		nasabah.FieldType:     nasabah.TypeSiswa,
		nasabah.FieldGender:   nasabah.GenderL,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik tidak valid")
}

// ── Constant correctness ─────────────────────────────────────────────────────

func TestConstants_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, nasabah.TypeSiswa)
	assert.NotEmpty(t, nasabah.TypeGuru)
	assert.NotEmpty(t, nasabah.TypeStaff)
	assert.NotEmpty(t, nasabah.TypeUmum)
	assert.NotEmpty(t, nasabah.GenderL)
	assert.NotEmpty(t, nasabah.GenderP)
	assert.NotEmpty(t, nasabah.StatusActive)
	assert.NotEmpty(t, nasabah.StatusInactive)
	assert.NotEmpty(t, nasabah.StatusBlacklisted)
}
