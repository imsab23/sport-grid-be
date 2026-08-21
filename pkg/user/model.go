package user

import (
	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrUserAlreadyExists = error.New("USR0000", "User already exists.")
	ErrUserNotFound      = error.New("USR0001", "User not found.")
)

const (
	UserTable = "users"
)

type User struct {
	ID            uuid.UUID `db:"id" json:"id"`
	FirstName     string    `db:"first_name" json:"first_name"`
	MiddleName    *string   `db:"middle_name" json:"middle_name"`
	LastName      string    `db:"last_name" json:"last_name"`
	Email         string    `db:"email" json:"email"`
	ContactNumber string    `db:"contact_number" json:"contact_number"`
	Password      string    `db:"password" json:"password"`
	CreatedBy     string    `db:"created_by" json:"created_by"`
	UpdatedBy     string    `db:"updated_by" json:"updated_by"`
	CreatedAt     string    `db:"created_at" json:"created_at"`
	UpdatedAt     string    `db:"updated_at" json:"updated_at"`
}

type CreateUserCommand struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	ContactNumber string  `json:"contact_number"`
	Password      string  `json:"password"`
}

type UpdateUserCommand struct {
	ID            uuid.UUID `json:"id"`
	FirstName     string    `json:"first_name"`
	MiddleName    *string   `json:"middle_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	ContactNumber string    `json:"contact_number"`
	UpdatedBy     string    `json:"updated_by"`
}

type SearchUserQuery struct {
	Email string `schema:"email"`
}

type SearchUserResult struct {
	Users []*User `json:"users"`
}
