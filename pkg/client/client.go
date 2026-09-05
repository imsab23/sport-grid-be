package client

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateClientCommand) (*Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Client, error)
	Search(ctx context.Context, query *SearchClientQuery) (*SearchClientResult, error)
	Suspend(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
}

type service struct {
	db db.DB
}

func NewService(database db.DB) (Service, error) {
	return &service{db: database}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateClientCommand) (*Client, error) {
	now := time.Now().UTC()
	c := &Client{
		Name:         cmd.Name,
		Slug:         strings.ToLower(cmd.Slug),
		Status:       StatusActive,
		ContactName:  cmd.ContactName,
		ContactEmail: cmd.ContactEmail,
		ContactPhone: cmd.ContactPhone,
		Address:      cmd.Address,
		CreatedBy:    cmd.CreatedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := helperdb.Create(ctx, s.db, ClientTable, c, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Client, error) {
	var c Client
	err := helperdb.GetByField(ctx, s.db, ClientTable, &Client{ID: id}, &c)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *service) Search(ctx context.Context, query *SearchClientQuery) (*SearchClientResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name", "slug"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Client](ctx, s.db, searchHelper.From{Table: ClientTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchClientResult{Clients: result.Items, Meta: &m}, nil
}

// clientStatusUpdate is a targeted struct so status changes only touch the relevant columns.
type clientStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Suspend(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, ClientTable, id, &clientStatusUpdate{
		Status:    StatusSuspended,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) Activate(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, ClientTable, id, &clientStatusUpdate{
		Status:    StatusActive,
		UpdatedAt: time.Now().UTC(),
	})
}
