package tournament

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrTournamentNotFound      = apperror.New("TRN0000", "Tournament not found.")
	ErrTournamentNameRequired  = apperror.New("TRN0001", "Tournament name is required.")
	ErrInvalidDateRange        = apperror.New("TRN0002", "End date must not be before start date.")
	ErrInvalidFormat           = apperror.New("TRN0003", "Invalid tournament format.")
	ErrSportIDRequired         = apperror.New("TRN0004", "Sport is required.")
	ErrClientIDRequired        = apperror.New("TRN0005", "Client is required.")
	ErrSportNotActive          = apperror.New("TRN0006", "Sport is not active.")
	ErrInvalidTournamentState  = apperror.New("TRN0007", "Invalid tournament state transition.")
	ErrCancelledReasonRequired = apperror.New("TRN0008", "Cancellation reason is required.")
)

const (
	TournamentTable = "tournaments"
)

type Status string

const (
	StatusDraft              Status = "DRAFT"
	StatusPublished          Status = "PUBLISHED"
	StatusRegistrationOpen   Status = "REGISTRATION_OPEN"
	StatusRegistrationClosed Status = "REGISTRATION_CLOSED"
	StatusOngoing            Status = "ONGOING"
	StatusCompleted          Status = "COMPLETED"
	StatusCancelled          Status = "CANCELLED"
)

type Format string

const (
	FormatSingleElimination Format = "SINGLE_ELIMINATION"
	FormatDoubleElimination Format = "DOUBLE_ELIMINATION"
	FormatRoundRobin        Format = "ROUND_ROBIN"
	FormatSwiss             Format = "SWISS"
)

func isValidFormat(f Format) bool {
	switch f {
	case FormatSingleElimination, FormatDoubleElimination, FormatRoundRobin, FormatSwiss:
		return true
	}
	return false
}

type Tournament struct {
	ID                  uuid.UUID  `db:"id"                     json:"id"`
	ClientID            uuid.UUID  `db:"client_id"              json:"client_id"`
	SportID             uuid.UUID  `db:"sport_id"               json:"sport_id"`
	Name                string     `db:"name"                   json:"name"`
	Description         *string    `db:"description"            json:"description,omitempty"`
	Venue               *string    `db:"venue"                  json:"venue,omitempty"`
	Address             *string    `db:"address"                json:"address,omitempty"`
	StartDate           time.Time  `db:"start_date"             json:"start_date"`
	EndDate             time.Time  `db:"end_date"               json:"end_date"`
	RegistrationOpenAt  *time.Time `db:"registration_open_at"   json:"registration_open_at,omitempty"`
	RegistrationCloseAt *time.Time `db:"registration_close_at"  json:"registration_close_at,omitempty"`
	RegistrationFee     float64    `db:"registration_fee"       json:"registration_fee"`
	MaxParticipants     *int       `db:"max_participants"       json:"max_participants,omitempty"`
	ContactName         *string    `db:"contact_name"           json:"contact_name,omitempty"`
	ContactEmail        *string    `db:"contact_email"          json:"contact_email,omitempty"`
	ContactPhone        *string    `db:"contact_phone"          json:"contact_phone,omitempty"`
	Rules               *string    `db:"rules"                  json:"rules,omitempty"`
	Terms               *string    `db:"terms"                  json:"terms,omitempty"`
	PaymentInstructions *string    `db:"payment_instructions"   json:"payment_instructions,omitempty"`
	Format              Format     `db:"format"                 json:"format"`
	Status              Status     `db:"status"                 json:"status"`
	CreatedBy           uuid.UUID  `db:"created_by"             json:"created_by"`
	CancelledReason     *string    `db:"cancelled_reason"       json:"cancelled_reason,omitempty"`
	CreatedAt           time.Time  `db:"created_at"             json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"             json:"updated_at"`
}

// CreateTournamentCommand.ClientID is bindable from JSON only for SuperAdmin;
// scoped roles have it overwritten by the handler from the authenticated identity.
type CreateTournamentCommand struct {
	ClientID            uuid.UUID  `json:"client_id"`
	SportID             uuid.UUID  `json:"sport_id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description"`
	Venue               *string    `json:"venue"`
	Address             *string    `json:"address"`
	StartDate           time.Time  `json:"start_date"`
	EndDate             time.Time  `json:"end_date"`
	RegistrationOpenAt  *time.Time `json:"registration_open_at"`
	RegistrationCloseAt *time.Time `json:"registration_close_at"`
	RegistrationFee     float64    `json:"registration_fee"`
	MaxParticipants     *int       `json:"max_participants"`
	ContactName         *string    `json:"contact_name"`
	ContactEmail        *string    `json:"contact_email"`
	ContactPhone        *string    `json:"contact_phone"`
	Rules               *string    `json:"rules"`
	Terms               *string    `json:"terms"`
	PaymentInstructions *string    `json:"payment_instructions"`
	Format              Format     `json:"format"`
	CreatedBy           uuid.UUID  `json:"-"`
}

func (cmd *CreateTournamentCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrTournamentNameRequired
	}
	if cmd.ClientID == uuid.Nil {
		return ErrClientIDRequired
	}
	if cmd.SportID == uuid.Nil {
		return ErrSportIDRequired
	}
	if !isValidFormat(cmd.Format) {
		return ErrInvalidFormat
	}
	if cmd.EndDate.Before(cmd.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}

type UpdateTournamentCommand struct {
	ID                  uuid.UUID
	Name                string     `db:"name"                  json:"name"`
	Description         *string    `db:"description"           json:"description"`
	Venue               *string    `db:"venue"                 json:"venue"`
	Address             *string    `db:"address"                json:"address"`
	StartDate           time.Time  `db:"start_date"             json:"start_date"`
	EndDate             time.Time  `db:"end_date"               json:"end_date"`
	RegistrationOpenAt  *time.Time `db:"registration_open_at"   json:"registration_open_at"`
	RegistrationCloseAt *time.Time `db:"registration_close_at"  json:"registration_close_at"`
	RegistrationFee     float64    `db:"registration_fee"       json:"registration_fee"`
	MaxParticipants     *int       `db:"max_participants"       json:"max_participants"`
	ContactName         *string    `db:"contact_name"           json:"contact_name"`
	ContactEmail        *string    `db:"contact_email"          json:"contact_email"`
	ContactPhone        *string    `db:"contact_phone"          json:"contact_phone"`
	Rules               *string    `db:"rules"                  json:"rules"`
	Terms               *string    `db:"terms"                  json:"terms"`
	PaymentInstructions *string    `db:"payment_instructions"   json:"payment_instructions"`
	UpdatedAt           time.Time  `db:"updated_at"             json:"updated_at"`
}

func (cmd *UpdateTournamentCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrTournamentNameRequired
	}
	if cmd.EndDate.Before(cmd.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}

type CancelTournamentCommand struct {
	ID     uuid.UUID
	Reason string `json:"reason"`
}

func (cmd *CancelTournamentCommand) Validate() error {
	if !validate.RequiredString(cmd.Reason) {
		return ErrCancelledReasonRequired
	}

	return nil
}

type SearchTournamentQuery struct {
	Search   string     `query:"search"`
	ClientID *uuid.UUID `db:"client_id" query:"client_id" empty:"skip"`
	SportID  *uuid.UUID `db:"sport_id"  query:"sport_id"  empty:"skip"`
	Status   *Status    `db:"status"    query:"status"    empty:"skip"`
	Meta     *meta.Meta
}

type SearchTournamentResult struct {
	Tournaments []*Tournament `json:"tournaments"`
	Meta        *meta.Meta    `json:"meta"`
}
