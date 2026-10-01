package court

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/pkg/tournament"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateCourtCommand) (*Court, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Court, error)
	Search(ctx context.Context, query *SearchCourtQuery) (*SearchCourtResult, error)
	Update(ctx context.Context, cmd *UpdateCourtCommand) (*Court, error)
	SetStatus(ctx context.Context, cmd *SetCourtStatusCommand) error
}

type service struct {
	db            db.DB
	tournamentSvc tournament.Service
}

func NewService(database db.DB, tournamentSvc tournament.Service) (Service, error) {
	return &service{db: database, tournamentSvc: tournamentSvc}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateCourtCommand) (*Court, error) {
	_, err := s.tournamentSvc.GetByID(ctx, cmd.TournamentID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	c := &Court{
		TournamentID: cmd.TournamentID,
		Name:         cmd.Name,
		Status:       StatusAvailable,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = helperdb.Create(ctx, s.db, CourtTable, c, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Court, error) {
	var c Court
	err := helperdb.GetByField(ctx, s.db, CourtTable, &Court{ID: id}, &c)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCourtNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *service) Search(ctx context.Context, query *SearchCourtQuery) (*SearchCourtResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Court](ctx, s.db, searchHelper.From{Table: CourtTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchCourtResult{Courts: result.Items, Meta: &m}, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCourtCommand) (*Court, error) {
	c, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	cmd.UpdatedAt = time.Now().UTC()
	err = helperdb.Update(ctx, s.db, CourtTable, c.ID, cmd)
	if err != nil {
		return nil, err
	}

	c.Name = cmd.Name
	c.UpdatedAt = cmd.UpdatedAt

	return c, nil
}

type courtStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) SetStatus(ctx context.Context, cmd *SetCourtStatusCommand) error {
	_, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	return helperdb.Update(ctx, s.db, CourtTable, cmd.ID, &courtStatusUpdate{
		Status:    cmd.Status,
		UpdatedAt: time.Now().UTC(),
	})
}
