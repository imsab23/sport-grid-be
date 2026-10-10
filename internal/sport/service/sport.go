package sportsvc

import (
	"context"
	"database/sql"
	"errors"
	"sport-grid-be/internal/sport"

	"github.com/google/uuid"
	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type service struct {
	db db.DB
}

func NewService(database db.DB) (sport.Service, error) {
	return &service{db: database}, nil
}

func (s *service) Create(ctx context.Context, cmd *sport.CreateSportCommand) (*sport.Sport, error) {
	existing, err := s.GetBySlug(ctx, cmd.Slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, sport.ErrSportAlreadyExists
	}

	sp := &sport.Sport{
		Name:        cmd.Name,
		Slug:        cmd.Slug,
		Description: cmd.Description,
		Status:      sport.Active,
	}

	_, err = helperdb.Create(ctx, s.db, sport.SportTable, sp, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*sport.Sport, error) {
	var result sport.Sport
	err := helperdb.GetByField(ctx, s.db, sport.SportTable, &sport.Sport{ID: id}, &result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sport.ErrSportNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*sport.Sport, error) {
	var result sport.Sport
	err := helperdb.GetByField(ctx, s.db, sport.SportTable, &sport.Sport{Slug: slug}, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *service) Search(ctx context.Context, query *sport.SearchSportQuery) (*sport.SearchSportResult, error) {
	params := searchHelper.Params{
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name", "slug"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*sport.Sport](ctx, s.db, searchHelper.From{Table: sport.SportTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &sport.SearchSportResult{Sports: result.Items, Meta: &m}, nil
}

func (s *service) Update(ctx context.Context, cmd *sport.UpdateSportCommand) (*sport.Sport, error) {
	sp, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	err = helperdb.Update(ctx, s.db, sport.SportTable, sp.ID, cmd)
	if err != nil {
		return nil, err
	}

	sp.Name = cmd.Name
	sp.Description = cmd.Description

	return sp, nil
}

func (s *service) UpdateStatus(ctx context.Context, cmd *sport.UpdateStatusCommand) error {
	result, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if cmd.Status == result.Status {
		return nil
	}

	return helperdb.Update(ctx, s.db, sport.SportTable, cmd.ID, cmd)
}
