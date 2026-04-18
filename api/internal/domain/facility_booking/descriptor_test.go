package facility_booking_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/facility_booking"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

func TestFacilityDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "facilities", (&facility_booking.FacilityDescriptor{}).TableName())
}
func TestFacilityDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &facility_booking.FacilityDescriptor{}
}
func TestFacilityDescriptor_DefaultRels(t *testing.T) {
	assert.Len(t, (&facility_booking.FacilityDescriptor{}).DefaultRels(), 0)
}

func validFacilityData() map[string]any {
	return map[string]any{facility_booking.FFieldName: "Aula Utama", facility_booking.FFieldType: facility_booking.FTypeHall}
}
func TestFacilityDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&facility_booking.FacilityDescriptor{}).Validate(validFacilityData()))
}
func TestFacilityDescriptor_Validate_MissingName(t *testing.T) {
	d := validFacilityData(); delete(d, facility_booking.FFieldName)
	assert.ErrorContains(t, (&facility_booking.FacilityDescriptor{}).Validate(d), "name wajib diisi")
}
func TestFacilityDescriptor_Validate_InvalidType(t *testing.T) {
	d := validFacilityData(); d[facility_booking.FFieldType] = "pool"
	assert.ErrorContains(t, (&facility_booking.FacilityDescriptor{}).Validate(d), "type tidak valid")
}

func TestBookingDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "facility_bookings", (&facility_booking.BookingDescriptor{}).TableName())
}
func TestBookingDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &facility_booking.BookingDescriptor{}
}
func TestBookingDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&facility_booking.BookingDescriptor{}).DefaultRels(), 2)
}
func TestBookingDescriptor_DefaultRels_Facility(t *testing.T) {
	rel, ok := (&facility_booking.BookingDescriptor{}).DefaultRels()[facility_booking.RelFacility]
	require.True(t, ok); assert.Equal(t, "facilities", rel.Domain); assert.True(t, rel.IsAutoload)
}

func validBookingData() map[string]any {
	return map[string]any{
		facility_booking.BFieldFacilityID: "00000000-0000-0000-0000-000000000001",
		facility_booking.BFieldRequesterType: facility_booking.BReqTeacher,
		facility_booking.BFieldRequesterID: "00000000-0000-0000-0000-000000000002",
		facility_booking.BFieldBookingDate: "2026-04-20", facility_booking.BFieldPurpose: "Rapat",
	}
}
func TestBookingDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&facility_booking.BookingDescriptor{}).Validate(validBookingData()))
}
func TestBookingDescriptor_Validate_MissingFacilityID(t *testing.T) {
	d := validBookingData(); delete(d, facility_booking.BFieldFacilityID)
	assert.ErrorContains(t, (&facility_booking.BookingDescriptor{}).Validate(d), "facility_id wajib diisi")
}
func TestBookingDescriptor_Validate_InvalidRequesterType(t *testing.T) {
	d := validBookingData(); d[facility_booking.BFieldRequesterType] = "parent"
	assert.ErrorContains(t, (&facility_booking.BookingDescriptor{}).Validate(d), "requester_type tidak valid")
}
func TestBookingDescriptor_Validate_MissingPurpose(t *testing.T) {
	d := validBookingData(); delete(d, facility_booking.BFieldPurpose)
	assert.ErrorContains(t, (&facility_booking.BookingDescriptor{}).Validate(d), "purpose wajib diisi")
}
