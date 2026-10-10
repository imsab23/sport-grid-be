package user

import (
	"errors"
	"testing"

	"sport-grid-be/internal/role"

	"github.com/google/uuid"
)

var (
	testActorClientID = uuid.MustParse("0199b88b-cd21-7f3a-bf6c-158dbfae85e7")
	testOtherClientID = uuid.MustParse("0199b88b-cd21-7f3a-bf6c-158dbfae85e8")
)

func validCreateUserCommand() CreateUserCommand {
	return CreateUserCommand{
		FirstName: "Alex",
		LastName:  "Morgan",
		Email:     "alex@example.com",
		Password:  "StrongPassword123!",
		Role:      role.Scorer,
		ClientID:  &testOtherClientID,
		Actor: role.Actor{
			Role:     role.ClientAdmin,
			ClientID: &testActorClientID,
		},
	}
}

func TestCreateUserCommandValidateScopesClientAdminToOwnClient(t *testing.T) {
	command := validCreateUserCommand()

	err := command.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if command.ClientID == nil || *command.ClientID != testActorClientID {
		t.Fatalf("ClientID = %v, want actor client %s", command.ClientID, testActorClientID)
	}
}

func TestCreateUserCommandValidateRejectsPrivilegeEscalation(t *testing.T) {
	testCases := []struct {
		name       string
		targetRole role.Role
	}{
		{name: "client admin", targetRole: role.ClientAdmin},
		{name: "super admin", targetRole: role.SuperAdmin},
		{name: "legacy tournament staff", targetRole: role.TournamentStaff},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			command := validCreateUserCommand()
			command.Role = testCase.targetRole

			err := command.Validate()
			if !errors.Is(err, ErrRoleAssignmentForbidden) && !errors.Is(err, ErrInvalidRole) {
				t.Fatalf("Validate() error = %v, want role rejection", err)
			}
		})
	}
}

func TestSearchUserQueryApplyActorScopeOverridesRequestedClient(t *testing.T) {
	query := SearchUserQuery{
		ClientID: &testOtherClientID,
		Actor: role.Actor{
			Role:     role.ClientAdmin,
			ClientID: &testActorClientID,
		},
	}

	err := query.ApplyActorScope()
	if err != nil {
		t.Fatalf("ApplyActorScope() error = %v", err)
	}

	if query.ClientID == nil || *query.ClientID != testActorClientID {
		t.Fatalf("ClientID = %v, want actor client %s", query.ClientID, testActorClientID)
	}
}

func TestSearchUserQueryApplyActorScopeRejectsOperationalStaff(t *testing.T) {
	query := SearchUserQuery{
		Actor: role.Actor{
			Role:     role.Scorer,
			ClientID: &testActorClientID,
		},
	}

	err := query.ApplyActorScope()
	if !errors.Is(err, ErrUserScopeForbidden) {
		t.Fatalf("ApplyActorScope() error = %v, want ErrUserScopeForbidden", err)
	}
}
