package transport

import (
	"sport-grid-be/internal/role"

	"github.com/google/uuid"
	platformauthz "github.com/imsab23/platform-be/pkg/security/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrForbidden       = apperror.New("AUTHZ0000", "You do not have permission to perform this action.")
	ErrNoClientContext = apperror.New("AUTHZ0001", "Your account is not associated with a client.")
)

func IsSuperAdmin(identityRecord *identity.Identity) bool {
	return platformauthz.HasRole(identityRecord, string(role.SuperAdmin))
}

func IsClientAdmin(identityRecord *identity.Identity) bool {
	return platformauthz.HasRole(identityRecord, string(role.ClientAdmin))
}

func ClientID(identityRecord *identity.Identity) (uuid.UUID, error) {
	if identityRecord == nil || identityRecord.ClientID == "" {
		return uuid.Nil, ErrNoClientContext
	}

	clientID, err := uuid.Parse(identityRecord.ClientID)
	if err != nil {
		return uuid.Nil, ErrNoClientContext
	}

	return clientID, nil
}

func ActorFromIdentity(identityRecord *identity.Identity) (role.Actor, error) {
	if identityRecord == nil || len(identityRecord.Roles) != 1 {
		return role.Actor{}, ErrForbidden
	}

	actorRole := role.Role(identityRecord.Roles[0])
	if !actorRole.IsKnown() {
		return role.Actor{}, ErrForbidden
	}

	actor := role.Actor{Role: actorRole}
	if actorRole == role.SuperAdmin {
		return actor, nil
	}

	clientID, err := ClientID(identityRecord)
	if err != nil {
		return role.Actor{}, err
	}

	actor.ClientID = &clientID
	return actor, nil
}

func RequireClientOwnership(identityRecord *identity.Identity, resourceClientID uuid.UUID) error {
	if IsSuperAdmin(identityRecord) {
		return nil
	}

	clientID, err := ClientID(identityRecord)
	if err != nil || clientID != resourceClientID {
		return ErrForbidden
	}

	return nil
}
