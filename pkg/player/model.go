package player

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

var (
	ErrPlayerProfileNotFound = apperror.New("PLY0000", "Player profile not found.")
)

const (
	PlayerTable = "players"
)

type Player struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	UserID                uuid.UUID  `db:"user_id" json:"user_id"`
	FirstName             string     `db:"first_name" json:"first_name"`
	MiddleName            *string    `db:"middle_name" json:"middle_name"`
	LastName              string     `db:"last_name" json:"last_name"`
	DateOfBirth           *time.Time `db:"date_of_birth" json:"date_of_birth,omitempty"`
	Gender                *string    `db:"gender" json:"gender,omitempty"`
	Phone                 *string    `db:"phone" json:"phone,omitempty"`
	ProfileImageURL       *string    `db:"profile_image_url" json:"profile_image_url,omitempty"`
	Address               *string    `db:"address" json:"address,omitempty"`
	EmergencyContactName  *string    `db:"emergency_contact_name" json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string    `db:"emergency_contact_phone" json:"emergency_contact_phone,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

type PlayerUpdate struct {
	FirstName             string     `db:"first_name"`
	MiddleName            *string    `db:"middle_name"`
	LastName              string     `db:"last_name"`
	DateOfBirth           *time.Time `db:"date_of_birth"`
	Gender                *string    `db:"gender"`
	Phone                 *string    `db:"phone"`
	Address               *string    `db:"address"`
	EmergencyContactName  *string    `db:"emergency_contact_name"`
	EmergencyContactPhone *string    `db:"emergency_contact_phone"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

type CreatePlayerCommand struct {
	UserID     uuid.UUID `json:"-"`
	FirstName  string    `json:"first_name"`
	MiddleName *string   `json:"middle_name"`
	LastName   string    `json:"last_name"`
	Phone      *string   `json:"phone"`
}

type UpdatePlayerCommand struct {
	UserID                uuid.UUID  `json:"-"`
	FirstName             string     `json:"first_name"`
	MiddleName            *string    `json:"middle_name"`
	LastName              string     `json:"last_name"`
	DateOfBirth           *time.Time `json:"date_of_birth"`
	Gender                *string    `json:"gender"`
	Phone                 *string    `json:"phone"`
	Address               *string    `json:"address"`
	EmergencyContactName  *string    `json:"emergency_contact_name"`
	EmergencyContactPhone *string    `json:"emergency_contact_phone"`
}

// PlayerSearchItem is the result row for a player search (players JOIN users).
type PlayerSearchItem struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	UserID     uuid.UUID `db:"user_id"     json:"user_id"`
	FirstName  string    `db:"first_name"  json:"first_name"`
	MiddleName *string   `db:"middle_name" json:"middle_name"`
	LastName   string    `db:"last_name"   json:"last_name"`
	Email      string    `db:"email"       json:"email"`
	Phone      *string   `db:"phone"       json:"phone,omitempty"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

type SearchPlayerQuery struct {
	Search *string   `json:"search"`
	Meta   meta.Meta `json:"meta"`
}

type SearchPlayerResult struct {
	Players []*PlayerSearchItem `json:"players"`
	Meta    meta.Meta           `json:"meta"`
}
