package role

import "github.com/google/uuid"

type Role string

const (
	SuperAdmin                 Role = "super_admin"
	ClientAdmin                Role = "client_admin"
	RegistrationManager        Role = "registration_manager"
	PaymentReviewer            Role = "payment_reviewer"
	ContentAnnouncementManager Role = "content_announcement_manager"
	Scorer                     Role = "scorer"
	Referee                    Role = "referee"

	// TournamentStaff is retained for existing accounts while callers migrate to
	// the approved operational roles. It cannot be assigned to new users.
	TournamentStaff Role = "tournament_staff"
)

type Actor struct {
	Role     Role
	ClientID *uuid.UUID
}

func (r Role) IsKnown() bool {
	switch r {
	case SuperAdmin,
		ClientAdmin,
		RegistrationManager,
		PaymentReviewer,
		ContentAnnouncementManager,
		Scorer,
		Referee,
		TournamentStaff:
		return true
	default:
		return false
	}
}

func (r Role) IsAssignable() bool {
	return r.IsKnown() && r != TournamentStaff
}

func (r Role) IsOperational() bool {
	switch r {
	case RegistrationManager, PaymentReviewer, ContentAnnouncementManager, Scorer, Referee:
		return true
	default:
		return false
	}
}

func CanAssign(actorRole Role, targetRole Role) bool {
	if !targetRole.IsAssignable() {
		return false
	}

	switch actorRole {
	case SuperAdmin:
		return true
	case ClientAdmin:
		return targetRole.IsOperational()
	default:
		return false
	}
}
