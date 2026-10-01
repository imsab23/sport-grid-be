package team

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrTeamNotFound             = apperror.New("TEA0000", "Team not found.")
	ErrTeamNameRequired         = apperror.New("TEA0001", "Team name is required.")
	ErrDivisionNotTeamBased     = apperror.New("TEA0002", "Division does not accept team registrations.")
	ErrNotRegisteredForDivision = apperror.New("TEA0003", "Player is not registered for this division.")
	ErrTeamNotActive            = apperror.New("TEA0004", "Team is not active.")
	ErrAlreadyOnTeam            = apperror.New("TEA0005", "Player is already on a team in this division.")
	ErrMemberNotFound           = apperror.New("TEA0006", "Team member not found.")
	ErrInvalidMemberRole        = apperror.New("TEA0007", "Invalid team member role.")
)

const (
	TeamTable       = "teams"
	TeamMemberTable = "team_members"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusDisbanded Status = "DISBANDED"
)

type MemberRole string

const (
	RoleCaptain MemberRole = "CAPTAIN"
	RoleMember  MemberRole = "MEMBER"
)

func isValidMemberRole(r MemberRole) bool {
	return r == RoleCaptain || r == RoleMember
}

type Team struct {
	ID             uuid.UUID  `db:"id"              json:"id"`
	TournamentID   uuid.UUID  `db:"tournament_id"   json:"tournament_id"`
	DivisionID     uuid.UUID  `db:"division_id"     json:"division_id"`
	RegistrationID *uuid.UUID `db:"registration_id" json:"registration_id,omitempty"`
	Name           string     `db:"name"            json:"name"`
	Status         Status     `db:"status"          json:"status"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
}

type TeamMember struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	TeamID    uuid.UUID  `db:"team_id"    json:"team_id"`
	PlayerID  uuid.UUID  `db:"player_id"  json:"player_id"`
	Role      MemberRole `db:"role"       json:"role"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// CreateTeamCommand.TournamentID/DivisionID are set by the handler from the URL path;
// CreatorPlayerID is set by the handler from the authenticated player identity.
type CreateTeamCommand struct {
	TournamentID    uuid.UUID `json:"-"`
	DivisionID      uuid.UUID `json:"-"`
	CreatorPlayerID uuid.UUID `json:"-"`
	Name            string    `json:"name"`
}

func (cmd *CreateTeamCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrTeamNameRequired
	}
	return nil
}

// AddTeamMemberCommand.TeamID is set by the handler from the URL path.
type AddTeamMemberCommand struct {
	TeamID   uuid.UUID  `json:"-"`
	PlayerID uuid.UUID  `json:"player_id"`
	Role     MemberRole `json:"role"`
}

func (cmd *AddTeamMemberCommand) Validate() error {
	if cmd.PlayerID == uuid.Nil {
		return ErrNotRegisteredForDivision
	}
	if cmd.Role == "" {
		cmd.Role = RoleMember
	}
	if !isValidMemberRole(cmd.Role) {
		return ErrInvalidMemberRole
	}
	return nil
}

type SearchTeamQuery struct {
	Search       string    `query:"search"`
	TournamentID uuid.UUID `db:"tournament_id" query:"-"`
	DivisionID   uuid.UUID `db:"division_id"   query:"-"`
	Status       *Status   `db:"status"        query:"status" empty:"skip"`
	Meta         *meta.Meta
}

type SearchTeamResult struct {
	Teams []*Team    `json:"teams"`
	Meta  *meta.Meta `json:"meta"`
}
