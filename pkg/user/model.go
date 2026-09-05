package user

import (
	"sport-grid-be/pkg/role"
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrUserAlreadyExists    = apperror.New("USR0000", "User already exists.")
	ErrUserNotFound         = apperror.New("USR0001", "User not found.")
	ErrFirstNameRequired    = apperror.New("USR0002", "First name is required.")
	ErrLastNameRequired     = apperror.New("USR0003", "Last name is required.")
	ErrEmailRequired        = apperror.New("USR0004", "Email is required.")
	ErrPasswordRequired     = apperror.New("USR0005", "Password is required.")
	ErrRoleRequired         = apperror.New("USR0006", "Role is required.")
	ErrClientIDRequired     = apperror.New("USR0007", "Client is required.")
	ErrInvalidContactNumber = apperror.New("USR0008", "Invalid contact number format.")
)

const (
	UserTable = "users"
)

type Status int

const (
	Active Status = iota + 1 // Active Status = 1
	Inactive
	Deleted
)

type User struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FirstName       string     `db:"first_name" json:"first_name"`
	MiddleName      *string    `db:"middle_name" json:"middle_name"`
	LastName        string     `db:"last_name" json:"last_name"`
	Email           string     `db:"email" json:"email"`
	ContactNumber   *string    `db:"contact_number" json:"contact_number,omitempty"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	Role            role.Role  `db:"role"      json:"role"`
	ClientID        *uuid.UUID `db:"client_id" json:"client_id,omitempty"`
	Status          Status     `db:"status" json:"status"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type SearchUserQuery struct {
	Search   string     `query:"search"`
	Email    *string    `db:"email" query:"email" empty:"skip"`
	Status   *Status    `db:"status" query:"status" empty:"skip"`
	ClientID *uuid.UUID `db:"client_id" query:"client_id" empty:"skip"`
	Role     *role.Role `db:"role" query:"role" empty:"skip"`
	Meta     *meta.Meta
}

type SearchUserResult struct {
	Users []*User    `json:"users"`
	Meta  *meta.Meta `json:"meta"`
}

type CreateUserCommand struct {
	FirstName     string     `json:"first_name"`
	MiddleName    *string    `json:"middle_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	ContactNumber *string    `json:"contact_number"`
	Password      string     `json:"password"`
	Role          role.Role  `json:"role"`
	ClientID      *uuid.UUID `json:"client_id"`
}

func (c *CreateUserCommand) Validate() error {
	if !validate.RequiredString(c.FirstName) {
		return ErrFirstNameRequired
	}

	if !validate.RequiredString(c.LastName) {
		return ErrLastNameRequired
	}

	if !validate.RequiredString(c.Email) {
		return ErrEmailRequired
	}

	if !validate.RequiredString(c.Password) {
		return ErrPasswordRequired
	}

	if !validate.RequiredString(string(c.Role)) {
		return ErrRoleRequired
	}

	if c.Role == role.SuperAdmin {
		c.ClientID = nil
	} else {
		if c.ClientID == nil || !validate.UUID(c.ClientID.String()) {
			return ErrClientIDRequired
		}
	}

	if c.ContactNumber != nil && !validate.Phone(*c.ContactNumber) {
		return ErrInvalidContactNumber
	}

	return nil
}

type UpdateUserCommand struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	ContactNumber *string `json:"contact_number"`
}

func (c *UpdateUserCommand) Validate() error {
	if !validate.RequiredString(c.FirstName) {
		return ErrFirstNameRequired
	}

	if !validate.RequiredString(c.LastName) {
		return ErrLastNameRequired
	}

	if c.ContactNumber != nil && !validate.Phone(*c.ContactNumber) {
		return ErrInvalidContactNumber
	}

	return nil
}
