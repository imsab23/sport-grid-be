package permission

import (
	"testing"

	"sport-grid-be/internal/role"
)

func TestAllPermissionsAreUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for _, permissionName := range All() {
		if permissionName == "" {
			t.Fatal("permission name must not be empty")
		}
		if _, exists := seen[permissionName]; exists {
			t.Fatalf("duplicate permission %q", permissionName)
		}
		seen[permissionName] = struct{}{}
	}
}

func TestForRoleUsesApprovedDefaults(t *testing.T) {
	testCases := []struct {
		name            string
		userRole        role.Role
		wantPermissions int
	}{
		{name: "super admin receives full catalog", userRole: role.SuperAdmin, wantPermissions: len(All())},
		{name: "client admin receives user management", userRole: role.ClientAdmin, wantPermissions: 2},
		{name: "unknown role fails closed", userRole: role.Role("unknown"), wantPermissions: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			permissions := ForRole(testCase.userRole)
			if len(permissions) != testCase.wantPermissions {
				t.Fatalf("permission count = %d, want %d", len(permissions), testCase.wantPermissions)
			}
		})
	}
}
