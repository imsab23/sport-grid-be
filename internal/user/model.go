package user

import (
	"time"

	"sport-grid-be/internal/role"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrUserAlreadyExists       = apperror.New("USR0000", "User already exists.")
	ErrUserNotFound            = apperror.New("USR0001", "User not found.")
	ErrFirstNameRequired       = apperror.New("USR0002", "First name is required.")
	ErrLastNameRequired        = apperror.New("USR0003", "Last name is required.")
	ErrEmailRequired           = apperror.New("USR0004", "Email is required.")
	ErrPasswordRequired        = apperror.New("USR0005", "Password is required.")
	ErrRoleRequired            = apperror.New("USR0006", "Role is required.")
	ErrClientIDRequired        = apperror.New("USR0007", "Client is required.")
	ErrInvalidContactNumber    = apperror.New("USR0008", "Invalid contact number format.")
	ErrOldPasswordRequired     = apperror.New("USR0009", "Old password is required.")
	ErrNewPasswordRequired     = apperror.New("USR0010", "New password is required.")
	ErrInvalidStatus           = apperror.New("USR0011", "Invalid status.")
	ErrInvalidRole             = apperror.New(apperror.CodeValidation, "Role is not supported.")
	ErrRoleAssignmentForbidden = apperror.New(apperror.CodeForbidden, "You cannot assign the requested role.")
	ErrUserScopeForbidden      = apperror.New(apperror.CodeForbidden, "You cannot access users outside your client.")
)

const UserTable = "users"

type Status int

const (
	Active Status = iota + 1
	Inactive
	Deleted
)

func (status Status) Validate() error {
	switch status {
	case Active, Inactive, Deleted:
		return nil
	default:
		return ErrInvalidStatus
	}
}

type User struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FirstName       string     `db:"first_name" json:"first_name"`
	MiddleName      *string    `db:"middle_name" json:"middle_name"`
	LastName        string     `db:"last_name" json:"last_name"`
	Email           string     `db:"email" json:"email"`
	ContactNumber   *string    `db:"contact_number" json:"contact_number,omitempty"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	Role            role.Role  `db:"role" json:"role"`
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
	Actor    role.Actor `db:"-" query:"-"`
}

func (query *SearchUserQuery) ApplyActorScope() error {
	if query.Actor.Role == role.SuperAdmin {
		return nil
	}

	if query.Actor.Role != role.ClientAdmin || query.Actor.ClientID == nil {
		return ErrUserScopeForbidden
	}

	query.ClientID = query.Actor.ClientID
	return nil
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
	Actor         role.Actor `json:"-"`
}

func (command *CreateUserCommand) Validate() error {
	if !validate.RequiredString(command.FirstName) {
		return ErrFirstNameRequired
	}

	if !validate.RequiredString(command.LastName) {
		return ErrLastNameRequired
	}

	if !validate.RequiredString(command.Email) {
		return ErrEmailRequired
	}

	if !validate.RequiredString(command.Password) {
		return ErrPasswordRequired
	}

	if !validate.RequiredString(string(command.Role)) {
		return ErrRoleRequired
	}

	if !command.Role.IsAssignable() {
		return ErrInvalidRole
	}

	if !role.CanAssign(command.Actor.Role, command.Role) {
		return ErrRoleAssignmentForbidden
	}

	err := command.applyActorClientScope()
	if err != nil {
		return err
	}

	if command.ContactNumber != nil && !validate.Phone(*command.ContactNumber) {
		return ErrInvalidContactNumber
	}

	return nil
}

func (command *CreateUserCommand) applyActorClientScope() error {
	if command.Role == role.SuperAdmin {
		command.ClientID = nil
		return nil
	}

	if command.Actor.Role == role.ClientAdmin {
		if command.Actor.ClientID == nil {
			return ErrClientIDRequired
		}

		command.ClientID = command.Actor.ClientID
		return nil
	}

	if command.ClientID == nil || !validate.CheckUUIDv7Format(command.ClientID.String()) {
		return ErrClientIDRequired
	}

	return nil
}

type UpdateUserCommand struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	ContactNumber *string `json:"contact_number"`
}

func (command *UpdateUserCommand) Validate() error {
	if !validate.RequiredString(command.FirstName) {
		return ErrFirstNameRequired
	}

	if !validate.RequiredString(command.LastName) {
		return ErrLastNameRequired
	}

	if command.ContactNumber != nil && !validate.Phone(*command.ContactNumber) {
		return ErrInvalidContactNumber
	}

	return nil
}

type UserLastLoginUpdate struct {
	LastLoginAt time.Time `db:"last_login_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type UpdateUserPasswordCommand struct {
	ID           uuid.UUID `json:"id"`
	OldPassword  string    `json:"old_password"`
	NewPassword  string    `json:"new_password"`
	PasswordHash string    `db:"password_hash"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (command *UpdateUserPasswordCommand) Validate() error {
	if !validate.RequiredString(command.OldPassword) {
		return ErrOldPasswordRequired
	}

	if !validate.RequiredString(command.NewPassword) {
		return ErrNewPasswordRequired
	}

	return nil
}

type UpdateUserStatusCommand struct {
	ID        uuid.UUID `json:"id"`
	Status    Status    `json:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (command *UpdateUserStatusCommand) Validate() error {
	if !validate.CheckUUIDv7Format(command.ID.String()) {
		return ErrUserNotFound
	}

	err := command.Status.Validate()
	if err != nil {
		return err
	}

	return nil
}
