package role_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/role"
)

func TestHasPermission_Exact(t *testing.T) {
	r := &role.Role{Permissions: []string{"students:read", "students:write"}}
	if !r.HasPermission("students:read") {
		t.Error("HasPermission should return true for exact match")
	}
}

func TestHasPermission_Missing(t *testing.T) {
	r := &role.Role{Permissions: []string{"students:read"}}
	if r.HasPermission("students:write") {
		t.Error("HasPermission should return false for missing permission")
	}
}

func TestHasPermission_Wildcard(t *testing.T) {
	r := &role.Role{Permissions: []string{"*"}}
	if !r.HasPermission("anything:at:all") {
		t.Error("HasPermission should return true for wildcard *")
	}
}

func TestHasPermission_Empty(t *testing.T) {
	r := &role.Role{Permissions: []string{}}
	if r.HasPermission("students:read") {
		t.Error("HasPermission should return false for empty permissions")
	}
}
