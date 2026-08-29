package client

import "github.com/google/uuid"

type Client struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
}
