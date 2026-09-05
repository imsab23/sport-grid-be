// Package authz provides tenant-ownership and role helpers shared across
// controllers and services. It never trusts client_id from request bodies —
// tenant scope always derives from the authenticated identity.
package authz

import (
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	platformauthz "github.com/imsab23/platform-be/pkg/security/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrForbidden       = apperror.New("AUTHZ0000", "You do not have permission to perform this action.")
	ErrNoClientContext = apperror.New("AUTHZ0001", "Your account is not associated with a client.")
)

// IsSuperAdmin reports whether id holds the SUPER_ADMIN role.
func IsSuperAdmin(id *identity.Identity) bool {
	return platformauthz.HasRole(id, string(role.SuperAdmin))
}

// IsClientAdmin reports whether id holds the CLIENT_ADMIN role.
func IsClientAdmin(id *identity.Identity) bool {
	return platformauthz.HasRole(id, string(role.ClientAdmin))
}

// ClientID parses the authenticated identity's client_id claim.
func ClientID(id *identity.Identity) (uuid.UUID, error) {
	if id == nil || id.ClientID == "" {
		return uuid.Nil, ErrNoClientContext
	}
	return uuid.Parse(id.ClientID)
}

// RequireClientOwnership denies access unless id is a Super Admin or belongs
// to resourceClientID. This is the enforcement point for tenant isolation.
func RequireClientOwnership(id *identity.Identity, resourceClientID uuid.UUID) error {
	if IsSuperAdmin(id) {
		return nil
	}
	cid, err := ClientID(id)
	if err != nil || cid != resourceClientID {
		return ErrForbidden
	}
	return nil
}
