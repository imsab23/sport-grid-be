package sport

import (
	"regexp"
	"strings"
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
	ErrSportInvalidStatus = apperror.New("SPT0003", "Invalid status.")
	ErrSportIDInvalid     = apperror.New("SPT0004", "Invalid sport ID.")
)

const (
	SportTable = "sports"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	multiHyphen     = regexp.MustCompile(`-+`)
)

type Status int

const (
	Active Status = iota + 1
	Inactive
)

func (s Status) Validate() error {
	switch s {
	case Active, Inactive:
		return nil
	default:
		return ErrSportInvalidStatus
	}
}

type Sport struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Slug        string    `db:"slug"        json:"slug"`
	Description *string   `db:"description" json:"description,omitempty"`
	Status      Status    `db:"status"      json:"status"`
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
	Name        string  `db:"name"        json:"name"`
	Description *string `db:"description" json:"description"`
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

type UpdateStatusCommand struct {
	ID     uuid.UUID
	Status Status `db:"status" json:"status"`
}

func (cmd *UpdateStatusCommand) Validate() error {
	if !validate.CheckUUIDv7Format(cmd.ID.String()) {
		return ErrSportIDInvalid
	}

	err := cmd.Status.Validate()
	if err != nil {
		return err
	}

	return nil
}

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))

	// Replace anything that isn't a-z or 0-9 with a hyphen.
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")

	// Remove duplicate hyphens.
	slug = multiHyphen.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens.
	slug = strings.Trim(slug, "-")

	return slug
}
