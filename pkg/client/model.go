package client

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

var (
	ErrClientNotFound  = apperror.New("CLI0000", "Client not found.")
	ErrClientSuspended = apperror.New("CLI0001", "Client is suspended and cannot perform this operation.")
)

const (
	ClientTable = "clients"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
)

type Client struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Slug         string    `db:"slug"          json:"slug"`
	Status       Status    `db:"status"        json:"status"`
	ContactName  *string   `db:"contact_name"  json:"contact_name,omitempty"`
	ContactEmail *string   `db:"contact_email" json:"contact_email,omitempty"`
	ContactPhone *string   `db:"contact_phone" json:"contact_phone,omitempty"`
	Address      *string   `db:"address"       json:"address,omitempty"`
	CreatedBy    uuid.UUID `db:"created_by"    json:"created_by"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type CreateClientCommand struct {
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ContactName  *string   `json:"contact_name"`
	ContactEmail *string   `json:"contact_email"`
	ContactPhone *string   `json:"contact_phone"`
	Address      *string   `json:"address"`
	CreatedBy    uuid.UUID `json:"-"`
}

type SearchClientQuery struct {
	Search string  `query:"search"`
	Status *Status `db:"status" query:"status" empty:"skip"`
	Meta   *meta.Meta
}

type SearchClientResult struct {
	Clients []*Client  `json:"clients"`
	Meta    *meta.Meta `json:"meta"`
}
