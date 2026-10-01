package court

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrCourtNotFound      = apperror.New("CRT0000", "Court not found.")
	ErrCourtNameRequired  = apperror.New("CRT0001", "Court name is required.")
	ErrInvalidCourtStatus = apperror.New("CRT0002", "Invalid court status.")
)

const (
	CourtTable = "courts"
)

type Status string

const (
	StatusAvailable   Status = "AVAILABLE"
	StatusOccupied    Status = "OCCUPIED"
	StatusMaintenance Status = "MAINTENANCE"
	StatusClosed      Status = "CLOSED"
)

func isValidStatus(st Status) bool {
	switch st {
	case StatusAvailable, StatusOccupied, StatusMaintenance, StatusClosed:
		return true
	}
	return false
}

type Court struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	TournamentID uuid.UUID `db:"tournament_id" json:"tournament_id"`
	Name         string    `db:"name"          json:"name"`
	Status       Status    `db:"status"        json:"status"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

// CreateCourtCommand.TournamentID is set by the handler from the URL path, never from the body.
type CreateCourtCommand struct {
	TournamentID uuid.UUID `json:"-"`
	Name         string    `json:"name"`
}

func (cmd *CreateCourtCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrCourtNameRequired
	}
	return nil
}

type UpdateCourtCommand struct {
	ID        uuid.UUID
	Name      string    `db:"name"       json:"name"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (cmd *UpdateCourtCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrCourtNameRequired
	}
	return nil
}

type SetCourtStatusCommand struct {
	ID     uuid.UUID
	Status Status `json:"status"`
}

func (cmd *SetCourtStatusCommand) Validate() error {
	if !isValidStatus(cmd.Status) {
		return ErrInvalidCourtStatus
	}
	return nil
}

type SearchCourtQuery struct {
	Search       string    `query:"search"`
	TournamentID uuid.UUID `db:"tournament_id" query:"-"`
	Status       *Status   `db:"status" query:"status" empty:"skip"`
	Meta         *meta.Meta
}

type SearchCourtResult struct {
	Courts []*Court   `json:"courts"`
	Meta   *meta.Meta `json:"meta"`
}
