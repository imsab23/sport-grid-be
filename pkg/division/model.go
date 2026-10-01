package division

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrDivisionNotFound       = apperror.New("DIV0000", "Division not found.")
	ErrDivisionNameRequired   = apperror.New("DIV0001", "Division name is required.")
	ErrInvalidParticipantType = apperror.New("DIV0002", "Invalid participant type.")
	ErrInvalidCapacity        = apperror.New("DIV0003", "Capacity must be at least 1.")
	ErrInvalidAgeRange        = apperror.New("DIV0004", "Max age must not be less than min age.")
	ErrTournamentNotEditable  = apperror.New("DIV0005", "Tournament is completed or cancelled and can no longer be edited.")
)

const (
	DivisionTable = "tournament_divisions"
)

type ParticipantType string

const (
	ParticipantIndividual ParticipantType = "INDIVIDUAL"
	ParticipantTeam       ParticipantType = "TEAM"
)

func isValidParticipantType(p ParticipantType) bool {
	return p == ParticipantIndividual || p == ParticipantTeam
}

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

type Division struct {
	ID              uuid.UUID       `db:"id"               json:"id"`
	TournamentID    uuid.UUID       `db:"tournament_id"    json:"tournament_id"`
	Name            string          `db:"name"             json:"name"`
	Gender          *string         `db:"gender"           json:"gender,omitempty"`
	MinAge          *int            `db:"min_age"          json:"min_age,omitempty"`
	MaxAge          *int            `db:"max_age"          json:"max_age,omitempty"`
	SkillLevel      *string         `db:"skill_level"      json:"skill_level,omitempty"`
	ParticipantType ParticipantType `db:"participant_type" json:"participant_type"`
	Capacity        *int            `db:"capacity"         json:"capacity,omitempty"`
	RegistrationFee *float64        `db:"registration_fee" json:"registration_fee,omitempty"`
	Format          *string         `db:"format"           json:"format,omitempty"`
	Status          Status          `db:"status"           json:"status"`
	BestOf          *int            `db:"best_of"          json:"best_of,omitempty"`
	TargetScore     *int            `db:"target_score"     json:"target_score,omitempty"`
	WinBy           *int            `db:"win_by"           json:"win_by,omitempty"`
	MaxScore        *int            `db:"max_score"        json:"max_score,omitempty"`
	CreatedAt       time.Time       `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"       json:"updated_at"`
}

// CreateDivisionCommand.TournamentID is set by the handler from the URL path, never from the body.
type CreateDivisionCommand struct {
	TournamentID    uuid.UUID       `json:"-"`
	Name            string          `json:"name"`
	Gender          *string         `json:"gender"`
	MinAge          *int            `json:"min_age"`
	MaxAge          *int            `json:"max_age"`
	SkillLevel      *string         `json:"skill_level"`
	ParticipantType ParticipantType `json:"participant_type"`
	Capacity        *int            `json:"capacity"`
	RegistrationFee *float64        `json:"registration_fee"`
	Format          *string         `json:"format"`
	BestOf          *int            `json:"best_of"`
	TargetScore     *int            `json:"target_score"`
	WinBy           *int            `json:"win_by"`
	MaxScore        *int            `json:"max_score"`
}

func (cmd *CreateDivisionCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrDivisionNameRequired
	}

	if cmd.ParticipantType == "" {
		cmd.ParticipantType = ParticipantIndividual
	}
	if !isValidParticipantType(cmd.ParticipantType) {
		return ErrInvalidParticipantType
	}

	if cmd.Capacity != nil && *cmd.Capacity < 1 {
		return ErrInvalidCapacity
	}

	if cmd.MinAge != nil && cmd.MaxAge != nil && *cmd.MaxAge < *cmd.MinAge {
		return ErrInvalidAgeRange
	}

	return nil
}

type UpdateDivisionCommand struct {
	ID              uuid.UUID
	Name            string          `db:"name"             json:"name"`
	Gender          *string         `db:"gender"           json:"gender"`
	MinAge          *int            `db:"min_age"          json:"min_age"`
	MaxAge          *int            `db:"max_age"          json:"max_age"`
	SkillLevel      *string         `db:"skill_level"      json:"skill_level"`
	ParticipantType ParticipantType `db:"participant_type" json:"participant_type"`
	Capacity        *int            `db:"capacity"         json:"capacity"`
	RegistrationFee *float64        `db:"registration_fee" json:"registration_fee"`
	Format          *string         `db:"format"           json:"format"`
	BestOf          *int            `db:"best_of"          json:"best_of"`
	TargetScore     *int            `db:"target_score"     json:"target_score"`
	WinBy           *int            `db:"win_by"           json:"win_by"`
	MaxScore        *int            `db:"max_score"        json:"max_score"`
	UpdatedAt       time.Time       `db:"updated_at"       json:"updated_at"`
}

func (cmd *UpdateDivisionCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrDivisionNameRequired
	}

	if cmd.ParticipantType == "" {
		cmd.ParticipantType = ParticipantIndividual
	}
	if !isValidParticipantType(cmd.ParticipantType) {
		return ErrInvalidParticipantType
	}

	if cmd.Capacity != nil && *cmd.Capacity < 1 {
		return ErrInvalidCapacity
	}

	if cmd.MinAge != nil && cmd.MaxAge != nil && *cmd.MaxAge < *cmd.MinAge {
		return ErrInvalidAgeRange
	}

	return nil
}

type SearchDivisionQuery struct {
	Search          string           `query:"search"`
	TournamentID    uuid.UUID        `db:"tournament_id" query:"-"`
	ParticipantType *ParticipantType `db:"participant_type" query:"participant_type" empty:"skip"`
	Status          *Status          `db:"status" query:"status" empty:"skip"`
	Meta            *meta.Meta
}

type SearchDivisionResult struct {
	Divisions []*Division `json:"divisions"`
	Meta      *meta.Meta  `json:"meta"`
}
