package rekening_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/rekening"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &rekening.Descriptor{}
	assert.Equal(t, "rekening", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &rekening.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "rekening should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()

	nasabahRel, ok := rels[rekening.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", nasabahRel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, nasabahRel.Type)
	assert.Equal(t, rekening.FieldNasabahID, nasabahRel.FK)
	assert.True(t, nasabahRel.IsAutoload, "nasabah should be autoloaded")
	assert.Equal(t, []string{"full_name", "no_nasabah", "type"}, nasabahRel.Fields)
}

func TestDescriptor_DefaultRels_ProdukAkadRelation(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()

	paRel, ok := rels[rekening.RelProdukAkad]
	require.True(t, ok, "should have produk_akad relation")

	assert.Equal(t, "produk_akad", paRel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, paRel.Type)
	assert.Equal(t, rekening.FieldProdukAkadID, paRel.FK)
	assert.True(t, paRel.IsAutoload, "produk_akad should be autoloaded")
	assert.Equal(t, []string{"name", "code", "type", "akad_type"}, paRel.Fields)
}

// ── Valid data — all status x balance combinations ───────────────────────────

func TestDescriptor_Validate_AllStatusCombinations(t *testing.T) {
	d := &rekening.Descriptor{}
	statuses := []string{rekening.StatusActive, rekening.StatusFrozen, rekening.StatusClosed}

	for _, status := range statuses {
		data := map[string]any{
			rekening.FieldNasabahID:    "00000000-0000-0000-0000-000000000001",
			rekening.FieldProdukAkadID: "00000000-0000-0000-0000-000000000002",
			rekening.FieldNoRekening:   "REK-001",
			rekening.FieldBalance:      float64(500000),
			rekening.FieldHoldBalance:  float64(0),
			rekening.FieldStatus:       status,
		}
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_StatusOptional(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldBalance:     float64(500000),
		rekening.FieldHoldBalance: float64(0),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "status is optional")
}

func TestDescriptor_Validate_ZeroBalance(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldBalance:     float64(0),
		rekening.FieldHoldBalance: float64(0),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "zero balance should be valid")
}

func TestDescriptor_Validate_VeryLargeBalance(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldBalance:     float64(999999999999),
		rekening.FieldHoldBalance: float64(0),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "very large balance should be valid")
}

// ── Balance boundary conditions ──────────────────────────────────────────────

func TestDescriptor_Validate_BalanceBoundaryConditions(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "positive balance",
			data:    map[string]any{rekening.FieldBalance: float64(100000)},
			wantErr: false,
		},
		{
			name:    "zero balance",
			data:    map[string]any{rekening.FieldBalance: float64(0)},
			wantErr: false,
		},
		{
			name:    "negative balance",
			data:    map[string]any{rekening.FieldBalance: float64(-1)},
			wantErr: true,
			errMsg:  "balance tidak boleh negatif",
		},
		{
			name:    "large negative balance",
			data:    map[string]any{rekening.FieldBalance: float64(-999999)},
			wantErr: true,
			errMsg:  "balance tidak boleh negatif",
		},
		{
			name:    "fractional balance",
			data:    map[string]any{rekening.FieldBalance: float64(100.50)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &rekening.Descriptor{}
			err := d.Validate(tt.data)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── Hold balance boundary conditions ─────────────────────────────────────────

func TestDescriptor_Validate_HoldBalanceBoundaryConditions(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "hold equals balance",
			data: map[string]any{
				rekening.FieldBalance:     float64(100000),
				rekening.FieldHoldBalance: float64(100000),
			},
			wantErr: false,
		},
		{
			name: "hold less than balance",
			data: map[string]any{
				rekening.FieldBalance:     float64(100000),
				rekening.FieldHoldBalance: float64(50000),
			},
			wantErr: false,
		},
		{
			name: "hold zero balance positive",
			data: map[string]any{
				rekening.FieldBalance:     float64(100000),
				rekening.FieldHoldBalance: float64(0),
			},
			wantErr: false,
		},
		{
			name: "hold exceeds balance by 1",
			data: map[string]any{
				rekening.FieldBalance:     float64(100000),
				rekening.FieldHoldBalance: float64(100001),
			},
			wantErr: true,
			errMsg:  "hold_balance tidak boleh lebih besar dari balance",
		},
		{
			name: "hold much larger than balance",
			data: map[string]any{
				rekening.FieldBalance:     float64(100),
				rekening.FieldHoldBalance: float64(999999),
			},
			wantErr: true,
			errMsg:  "hold_balance tidak boleh lebih besar dari balance",
		},
		{
			name: "zero balance zero hold",
			data: map[string]any{
				rekening.FieldBalance:     float64(0),
				rekening.FieldHoldBalance: float64(0),
			},
			wantErr: false,
		},
		{
			name: "zero balance positive hold",
			data: map[string]any{
				rekening.FieldBalance:     float64(0),
				rekening.FieldHoldBalance: float64(1),
			},
			wantErr: true,
			errMsg:  "hold_balance tidak boleh lebih besar dari balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &rekening.Descriptor{}
			err := d.Validate(tt.data)
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

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"active is valid", rekening.StatusActive, false},
		{"frozen is valid", rekening.StatusFrozen, false},
		{"closed is valid", rekening.StatusClosed, false},
		{"unknown is invalid", "unknown", true},
		{"pending is invalid", "pending", true},
		{"ACTIVE uppercase is invalid", "ACTIVE", true},
		{"random string is invalid", "xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &rekening.Descriptor{}
			data := map[string]any{
				rekening.FieldStatus: tt.status,
			}
			err := d.Validate(data)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "status tidak valid")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── Empty/nil data map ──────────────────────────────────────────────────────

func TestDescriptor_Validate_NilData(t *testing.T) {
	d := &rekening.Descriptor{}
	// rekening has no required fields — nil should pass
	err := d.Validate(nil)
	assert.NoError(t, err, "rekening has no required fields; nil should pass")
}

func TestDescriptor_Validate_EmptyData(t *testing.T) {
	d := &rekening.Descriptor{}
	err := d.Validate(map[string]any{})
	assert.NoError(t, err, "rekening has no required fields; empty map should pass")
}

// ── Combined: negative balance + hold exceeds balance ────────────────────────

func TestDescriptor_Validate_NegativeBalanceCheckedBeforeHold(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldBalance:     float64(-1000),
		rekening.FieldHoldBalance: float64(5000),
	}
	err := d.Validate(data)
	require.Error(t, err)
	// Balance check runs first, so that error should be returned.
	assert.Contains(t, err.Error(), "balance tidak boleh negatif")
}

// ── Only hold_balance provided (no balance) ─────────────────────────────────

func TestDescriptor_Validate_HoldWithoutBalance(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldHoldBalance: float64(5000),
	}
	// If balance is not provided (not float64), hold check should pass
	// because both holdOK && balanceOK must be true for the comparison.
	err := d.Validate(data)
	assert.NoError(t, err, "hold_balance alone without balance should be OK (can't compare)")
}

// ── Wrong types for fields ──────────────────────────────────────────────────

func TestDescriptor_Validate_BalanceAsString(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldBalance: "100000", // string, not float64
	}
	err := d.Validate(data)
	assert.NoError(t, err, "non-float64 balance should be ignored by type assertion")
}

func TestDescriptor_Validate_StatusAsInt(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		rekening.FieldStatus: 123, // int, not string
	}
	err := d.Validate(data)
	assert.NoError(t, err, "non-string status should be treated as empty string")
}

// ── Constant correctness ─────────────────────────────────────────────────────

func TestConstants_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, rekening.StatusActive)
	assert.NotEmpty(t, rekening.StatusFrozen)
	assert.NotEmpty(t, rekening.StatusClosed)
	assert.NotEmpty(t, rekening.RelNasabah)
	assert.NotEmpty(t, rekening.RelProdukAkad)
}

func TestConstants_FieldNames(t *testing.T) {
	assert.Equal(t, "nasabah_id", rekening.FieldNasabahID)
	assert.Equal(t, "produk_akad_id", rekening.FieldProdukAkadID)
	assert.Equal(t, "no_rekening", rekening.FieldNoRekening)
	assert.Equal(t, "balance", rekening.FieldBalance)
	assert.Equal(t, "hold_balance", rekening.FieldHoldBalance)
	assert.Equal(t, "status", rekening.FieldStatus)
	assert.Equal(t, "open_date", rekening.FieldOpenDate)
	assert.Equal(t, "close_date", rekening.FieldCloseDate)
}
