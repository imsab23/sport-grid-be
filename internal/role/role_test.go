package role

import "testing"

func TestCanAssignEnforcesApprovedRoleHierarchy(t *testing.T) {
	testCases := []struct {
		name            string
		actorRole       Role
		targetRole      Role
		expectedAllowed bool
	}{
		{name: "super admin can assign client admin", actorRole: SuperAdmin, targetRole: ClientAdmin, expectedAllowed: true},
		{name: "super admin can assign operational role", actorRole: SuperAdmin, targetRole: Scorer, expectedAllowed: true},
		{name: "client admin can assign operational role", actorRole: ClientAdmin, targetRole: PaymentReviewer, expectedAllowed: true},
		{name: "client admin cannot assign client admin", actorRole: ClientAdmin, targetRole: ClientAdmin, expectedAllowed: false},
		{name: "client admin cannot assign super admin", actorRole: ClientAdmin, targetRole: SuperAdmin, expectedAllowed: false},
		{name: "operational staff cannot assign roles", actorRole: Scorer, targetRole: Referee, expectedAllowed: false},
		{name: "legacy tournament staff cannot be assigned", actorRole: SuperAdmin, targetRole: TournamentStaff, expectedAllowed: false},
		{name: "unknown role cannot be assigned", actorRole: SuperAdmin, targetRole: Role("unknown"), expectedAllowed: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualAllowed := CanAssign(testCase.actorRole, testCase.targetRole)
			if actualAllowed != testCase.expectedAllowed {
				t.Fatalf("CanAssign(%q, %q) = %t, want %t", testCase.actorRole, testCase.targetRole, actualAllowed, testCase.expectedAllowed)
			}
		})
	}
}
