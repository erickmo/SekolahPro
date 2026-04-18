package asset_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/asset"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

func TestAssetDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "assets", (&asset.AssetDescriptor{}).TableName())
}
func TestAssetDescriptor_ImplementsInterface(t *testing.T) { var _ vernon.DomainDescriptor = &asset.AssetDescriptor{} }
func TestAssetDescriptor_DefaultRels_Count(t *testing.T)   { assert.Len(t, (&asset.AssetDescriptor{}).DefaultRels(), 1) }
func TestAssetDescriptor_DefaultRels_Room(t *testing.T) {
	rel, ok := (&asset.AssetDescriptor{}).DefaultRels()[asset.RelRoom]
	require.True(t, ok); assert.Equal(t, "class_rooms", rel.Domain); assert.True(t, rel.IsAutoload)
}

func validAssetData() map[string]any {
	return map[string]any{asset.FieldAssetCode: "AST-001", asset.FieldName: "Proyektor", asset.FieldCategory: asset.CatPeralatan}
}
func TestAssetDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&asset.AssetDescriptor{}).Validate(validAssetData()))
}
func TestAssetDescriptor_Validate_MissingCode(t *testing.T) {
	d := validAssetData(); delete(d, asset.FieldAssetCode)
	assert.ErrorContains(t, (&asset.AssetDescriptor{}).Validate(d), "asset_code wajib diisi")
}
func TestAssetDescriptor_Validate_InvalidCategory(t *testing.T) {
	d := validAssetData(); d[asset.FieldCategory] = "furniture"
	assert.ErrorContains(t, (&asset.AssetDescriptor{}).Validate(d), "category tidak valid")
}

func TestMaintenanceDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "asset_maintenances", (&asset.MaintenanceDescriptor{}).TableName())
}
func TestMaintenanceDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &asset.MaintenanceDescriptor{}
}
func TestMaintenanceDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&asset.MaintenanceDescriptor{}).DefaultRels(), 1)
}
func TestMaintenanceDescriptor_DefaultRels_Asset(t *testing.T) {
	rel, ok := (&asset.MaintenanceDescriptor{}).DefaultRels()[asset.RelAsset]
	require.True(t, ok); assert.Equal(t, "assets", rel.Domain); assert.True(t, rel.IsAutoload)
}

func validMaintenanceData() map[string]any {
	return map[string]any{
		asset.MtFieldAssetID: "00000000-0000-0000-0000-000000000001",
		asset.MtFieldType: asset.MtPreventive, asset.MtFieldDescription: "Service rutin",
		asset.MtFieldStartDate: "2026-04-01",
	}
}
func TestMaintenanceDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&asset.MaintenanceDescriptor{}).Validate(validMaintenanceData()))
}
func TestMaintenanceDescriptor_Validate_MissingAssetID(t *testing.T) {
	d := validMaintenanceData(); delete(d, asset.MtFieldAssetID)
	assert.ErrorContains(t, (&asset.MaintenanceDescriptor{}).Validate(d), "asset_id wajib diisi")
}
func TestMaintenanceDescriptor_Validate_InvalidType(t *testing.T) {
	d := validMaintenanceData(); d[asset.MtFieldType] = "upgrade"
	assert.ErrorContains(t, (&asset.MaintenanceDescriptor{}).Validate(d), "type tidak valid")
}
