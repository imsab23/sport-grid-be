package match

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

var (
	ErrMatchNotFound      = apperror.New("MAT0000", "Match not found.")
	ErrMatchNotEditable   = apperror.New("MAT0001", "Match cannot be modified in its current state.")
	ErrInvalidWinnerSide  = apperror.New("MAT0002", "Winner side must be 1 or 2.")
	ErrInvalidResultType  = apperror.New("MAT0003", "Invalid match result type.")
	ErrGameNumberRequired = apperror.New("MAT0004", "Game number must be at least 1.")
	ErrMatchAlreadyFinal  = apperror.New("MAT0005", "Match has already been finalized.")
)

const (
	MatchTable            = "matches"
	MatchParticipantTable = "match_participants"
	MatchGameTable        = "match_games"
)

type Status string

const (
	StatusScheduled  Status = "SCHEDULED"
	StatusReady      Status = "READY"
	StatusInProgress Status = "IN_PROGRESS"
	StatusCompleted  Status = "COMPLETED"
	StatusCancelled  Status = "CANCELLED"
	StatusPostponed  Status = "POSTPONED"
)

type ResultType string

const (
	ResultNormalWin        ResultType = "NORMAL_WIN"
	ResultWalkover         ResultType = "WALKOVER"
	ResultForfeit          ResultType = "FORFEIT"
	ResultDisqualification ResultType = "DISQUALIFICATION"
	ResultBye              ResultType = "BYE"
)

func isValidResultType(r ResultType) bool {
	switch r {
	case ResultNormalWin, ResultWalkover, ResultForfeit, ResultDisqualification, ResultBye:
		return true
	}
	return false
}

type GameStatus string

const (
	GamePending   GameStatus = "PENDING"
	GameCompleted GameStatus = "COMPLETED"
)

type Match struct {
	ID                   uuid.UUID   `db:"id"                      json:"id"`
	TournamentID         uuid.UUID   `db:"tournament_id"           json:"tournament_id"`
	DivisionID           uuid.UUID   `db:"division_id"             json:"division_id"`
	BracketID            *uuid.UUID  `db:"bracket_id"              json:"bracket_id,omitempty"`
	BracketNodeID        *uuid.UUID  `db:"bracket_node_id"         json:"bracket_node_id,omitempty"`
	Round                int         `db:"round"                   json:"round"`
	MatchNumber          int         `db:"match_number"            json:"match_number"`
	CourtID              *uuid.UUID  `db:"court_id"                json:"court_id,omitempty"`
	ScheduledAt          *time.Time  `db:"scheduled_at"            json:"scheduled_at,omitempty"`
	Status               Status      `db:"status"                  json:"status"`
	ResultType           *ResultType `db:"result_type"            json:"result_type,omitempty"`
	WinnerRegistrationID *uuid.UUID  `db:"winner_registration_id" json:"winner_registration_id,omitempty"`
	WinnerTeamID         *uuid.UUID  `db:"winner_team_id"         json:"winner_team_id,omitempty"`
	Notes                *string     `db:"notes"                  json:"notes,omitempty"`
	FinalizedBy          *uuid.UUID  `db:"finalized_by"           json:"finalized_by,omitempty"`
	FinalizedAt          *time.Time  `db:"finalized_at"          json:"finalized_at,omitempty"`
	CreatedAt            time.Time   `db:"created_at"             json:"created_at"`
	UpdatedAt            time.Time   `db:"updated_at"             json:"updated_at"`
}

type MatchParticipant struct {
	ID             uuid.UUID  `db:"id"              json:"id"`
	MatchID        uuid.UUID  `db:"match_id"        json:"match_id"`
	RegistrationID *uuid.UUID `db:"registration_id" json:"registration_id,omitempty"`
	TeamID         *uuid.UUID `db:"team_id"         json:"team_id,omitempty"`
	Side           int        `db:"side"            json:"side"`
	IsWinner       bool       `db:"is_winner"       json:"is_winner"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
}

type MatchGame struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	MatchID    uuid.UUID  `db:"match_id"    json:"match_id"`
	GameNumber int        `db:"game_number" json:"game_number"`
	ScoreSide1 int        `db:"score_side1" json:"score_side1"`
	ScoreSide2 int        `db:"score_side2" json:"score_side2"`
	IsTiebreak bool       `db:"is_tiebreak" json:"is_tiebreak"`
	Status     GameStatus `db:"status"      json:"status"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
}

// ScheduleMatchCommand.MatchID is set by the handler from the URL path.
type ScheduleMatchCommand struct {
	MatchID     uuid.UUID  `json:"-"`
	CourtID     *uuid.UUID `json:"court_id"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// RecordGameCommand.MatchID is set by the handler from the URL path.
type RecordGameCommand struct {
	MatchID    uuid.UUID `json:"-"`
	GameNumber int       `json:"game_number"`
	ScoreSide1 int       `json:"score_side1"`
	ScoreSide2 int       `json:"score_side2"`
	IsTiebreak bool      `json:"is_tiebreak"`
}

func (cmd *RecordGameCommand) Validate() error {
	if cmd.GameNumber < 1 {
		return ErrGameNumberRequired
	}
	return nil
}

// FinalizeMatchCommand.MatchID/FinalizedBy are set by the handler.
type FinalizeMatchCommand struct {
	MatchID     uuid.UUID  `json:"-"`
	FinalizedBy uuid.UUID  `json:"-"`
	WinnerSide  int        `json:"winner_side"`
	ResultType  ResultType `json:"result_type"`
	Notes       *string    `json:"notes"`
}

func (cmd *FinalizeMatchCommand) Validate() error {
	if cmd.WinnerSide != 1 && cmd.WinnerSide != 2 {
		return ErrInvalidWinnerSide
	}
	if !isValidResultType(cmd.ResultType) {
		return ErrInvalidResultType
	}
	return nil
}

type CancelMatchCommand struct {
	MatchID uuid.UUID
	Reason  *string `json:"reason"`
}

type SearchMatchQuery struct {
	TournamentID uuid.UUID  `db:"tournament_id" query:"-"`
	DivisionID   *uuid.UUID `db:"division_id"  query:"division_id" empty:"skip"`
	Status       *Status    `db:"status"        query:"status"      empty:"skip"`
	Meta         *meta.Meta
}

type SearchMatchResult struct {
	Matches []*Match   `json:"matches"`
	Meta    *meta.Meta `json:"meta"`
}
