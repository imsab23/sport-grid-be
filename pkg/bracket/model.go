package bracket

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrBracketNotFound          = apperror.New("BRK0000", "Bracket not found.")
	ErrUnsupportedBracketFormat = apperror.New("BRK0001", "Only single-elimination brackets are currently supported.")
	ErrTeamDivisionNotSupported = apperror.New("BRK0002", "Bracket generation for team-based divisions is not yet supported.")
	ErrNoSeedingsFound          = apperror.New("BRK0003", "Division must be seeded before a bracket can be generated.")
	ErrInsufficientSeeds        = apperror.New("BRK0004", "At least 2 seeded registrations are required to generate a bracket.")
	ErrResetReasonRequired      = apperror.New("BRK0005", "Reset reason is required.")
)

const (
	BracketTable     = "brackets"
	BracketNodeTable = "bracket_nodes"
)

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusReset  Status = "RESET"
)

type Bracket struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	TournamentID uuid.UUID `db:"tournament_id" json:"tournament_id"`
	DivisionID   uuid.UUID `db:"division_id"   json:"division_id"`
	Format       string    `db:"format"         json:"format"`
	Status       Status    `db:"status"         json:"status"`
	GeneratedBy  uuid.UUID `db:"generated_by"   json:"generated_by"`
	GeneratedAt  time.Time `db:"generated_at"   json:"generated_at"`
	ResetReason  *string   `db:"reset_reason"   json:"reset_reason,omitempty"`
	CreatedAt    time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updated_at"`
}

// BracketNode is a single slot in the bracket tree. Round 1 holds the seeded leaves;
// each subsequent round holds the winner slot a node's occupant advances to via NextNodeID.
type BracketNode struct {
	ID             uuid.UUID  `db:"id"              json:"id"`
	BracketID      uuid.UUID  `db:"bracket_id"      json:"bracket_id"`
	Round          int        `db:"round"           json:"round"`
	Position       int        `db:"position"        json:"position"`
	RegistrationID *uuid.UUID `db:"registration_id" json:"registration_id,omitempty"`
	TeamID         *uuid.UUID `db:"team_id"         json:"team_id,omitempty"`
	IsBye          bool       `db:"is_bye"          json:"is_bye"`
	NextNodeID     *uuid.UUID `db:"next_node_id"    json:"next_node_id,omitempty"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
}

// GenerateBracketCommand.DivisionID is set by the handler from the URL path.
type GenerateBracketCommand struct {
	DivisionID  uuid.UUID `json:"-"`
	GeneratedBy uuid.UUID `json:"-"`
}

type ResetBracketCommand struct {
	ID          uuid.UUID
	Reason      string    `json:"reason"`
	GeneratedBy uuid.UUID `json:"-"`
}

func (cmd *ResetBracketCommand) Validate() error {
	if cmd.Reason == "" {
		return ErrResetReasonRequired
	}
	return nil
}
