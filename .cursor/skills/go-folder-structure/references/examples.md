# Examples

A complete generic `client` domain across files. Platform names follow the platform README; confirm in the platform skill. For full style details see `go-code-standard`.

## Tree

```text
internal/client/
├── client.go
├── model.go
└── service/
    └── client_service.go
transport/http/handler/client.go
transport/http/routes.go
migration/002_create_clients.sql
cmd/app/main.go
```

## `migration/002_create_clients.sql`

```sql
CREATE TABLE clients (
    id         UUID PRIMARY KEY,
    name       TEXT        NOT NULL,
    status     SMALLINT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX clients_name_unique ON clients (LOWER(name));
```

## `internal/client/client.go`

```go
package client

import (
	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrClientNotFound         = apperror.New("CLI0000", "Client not found.")
	ErrClientAlreadySuspended = apperror.New("CLI0001", "Client is already suspended.")
	ErrClientNameRequired     = apperror.New("CLI0002", "Client name is required.")
	ErrInvalidStatus          = apperror.New("CLI0003", "Invalid client status.")
)
```

## `internal/client/model.go`

```go
package client

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status int

const (
	StatusActive Status = iota + 1
	StatusSuspended
)

func (s Status) Validate() error {
	switch s {
	case StatusActive, StatusSuspended:
		return nil
	default:
		return ErrInvalidStatus
	}
}

type Client struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Status    Status    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateClientCommand struct {
	Name string `json:"name"`
}

func (c *CreateClientCommand) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrClientNameRequired
	}

	return nil
}
```

## `internal/client/service/client_service.go`

```go
package clientsvc

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"

	"github.com/your-org/project/internal/client"
)

const (
	selectClientByIDQuery = `SELECT id, name, status, created_at, updated_at FROM clients WHERE id = ?`
	insertClientQuery     = `INSERT INTO clients (id, name, status) VALUES (:id, :name, :status)`
	suspendClientQuery    = `UPDATE clients SET status = ?, updated_at = NOW() WHERE id = ?`
)

type ClientService struct {
	database db.DB
}

func NewClientService(database db.DB) *ClientService {
	return &ClientService{database: database}
}

func (s *ClientService) GetByID(ctx context.Context, clientID uuid.UUID) (*client.Client, error) {
	var clientRecord client.Client

	err := s.database.Get(ctx, &clientRecord, selectClientByIDQuery, clientID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, client.ErrClientNotFound
	}

	if err != nil {
		return nil, err
	}

	return &clientRecord, nil
}

func (s *ClientService) Create(ctx context.Context, command client.CreateClientCommand) (*client.Client, error) {
	err := command.Validate()

	if err != nil {
		return nil, err
	}

	createdClient := client.Client{
		ID:     uuid.New(),
		Name:   strings.TrimSpace(command.Name),
		Status: client.StatusActive,
	}

	_, err = s.database.NamedExec(ctx, insertClientQuery, createdClient)

	if err != nil {
		return nil, err
	}

	return &createdClient, nil
}

func (s *ClientService) Suspend(ctx context.Context, clientID uuid.UUID) error {
	foundClient, err := s.GetByID(ctx, clientID)

	if err != nil {
		return err
	}

	if foundClient.Status == client.StatusSuspended {
		return client.ErrClientAlreadySuspended
	}

	_, err = s.database.Exec(ctx, suspendClientQuery, client.StatusSuspended, clientID)

	if err != nil {
		return err
	}

	return nil
}
```

## `transport/http/handler/client.go`

```go
package handler

import (
	"github.com/google/uuid"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	apperror "github.com/imsab23/platform-be/pkg/util/error"

	"github.com/your-org/project/internal/client"
	clientsvc "github.com/your-org/project/internal/client/service"
)

type ClientHandler struct {
	clientService *clientsvc.ClientService
}

func NewClientHandler(clientService *clientsvc.ClientService) *ClientHandler {
	return &ClientHandler{clientService: clientService}
}

func (h *ClientHandler) Create(c *router.Ctx) error {
	var command client.CreateClientCommand

	err := c.Bind(&command)

	if err != nil {
		return apperror.BadRequest("invalid request body")
	}

	createdClient, err := h.clientService.Create(c.Context(), command)

	if err != nil {
		return err
	}

	response.Created(c.ResponseWriter(), "client created", createdClient)

	return nil
}

func (h *ClientHandler) Suspend(c *router.Ctx) error {
	clientID, err := uuid.Parse(c.Param("id"))

	if err != nil {
		return apperror.BadRequest("invalid client id")
	}

	err = h.clientService.Suspend(c.Context(), clientID)

	if err != nil {
		return err
	}

	return c.NoContent()
}
```

## `transport/http/routes.go` (excerpt)

```go
func (s *Server) setupRoutes() {
	s.router.Group("/api/v1", func(g router.Router) {
		g.Use(chirtr.WrapNetHTTPMiddleware(authmw.New(s.verifier)))

		g.POST("/clients", s.clientHandler.Create)
		g.POST("/clients/{id}/suspend", s.clientHandler.Suspend)
	})
}
```

## What this example shows

- The model package has no SQL and imports nothing from services or transport.
- The service owns its SQL and its transaction scope, uses the platform database directly, and translates `sql.ErrNoRows` into the domain error.
- The handler only binds, parses, calls, and returns; the router maps the returned application error.
- `main.go` constructs `clientsvc.NewClientService(database)` and `handler.NewClientHandler(clientService)` and passes them to the server.
