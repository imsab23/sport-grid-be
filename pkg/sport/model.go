package sport

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrSportNotFound      = apperror.New("SPT0000", "Sport not found.")
	ErrSportAlreadyExists = apperror.New("SPT0001", "Sport already exists.")
	ErrSportNameRequired  = apperror.New("SPT0002", "Sport name is required.")
)

const (
	SportTable = "sports"
)

type Sport struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Slug        string    `db:"slug"        json:"slug"`
	Description *string   `db:"description" json:"description,omitempty"`
	IsActive    bool      `db:"is_active"   json:"is_active"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

type CreateSportCommand struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

func (cmd *CreateSportCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrSportNameRequired
	}

	cmd.Slug = generateSlug(cmd.Name)

	return nil
}

type UpdateSportCommand struct {
	ID          uuid.UUID
	Name        string    `db:"name"        json:"name"`
	Description *string   `db:"description" json:"description"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

func (cmd *UpdateSportCommand) Validate() error {
	if !validate.RequiredString(cmd.Name) {
		return ErrSportNameRequired
	}

	return nil
}

type SearchSportQuery struct {
	Search   string `query:"search"`
	IsActive *bool  `db:"is_active" query:"is_active" empty:"skip"`
	Meta     *meta.Meta
}

type SearchSportResult struct {
	Sports []*Sport   `json:"sports"`
	Meta   *meta.Meta `json:"meta"`
}
