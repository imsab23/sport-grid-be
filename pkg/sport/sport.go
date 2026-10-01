package sport

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateSportCommand) (*Sport, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Sport, error)
	GetBySlug(ctx context.Context, slug string) (*Sport, error)
	Search(ctx context.Context, query *SearchSportQuery) (*SearchSportResult, error)
	Update(ctx context.Context, cmd *UpdateSportCommand) (*Sport, error)
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
}

type service struct {
	db db.DB
}

func NewService(database db.DB) (Service, error) {
	return &service{db: database}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateSportCommand) (*Sport, error) {
	existing, err := s.GetBySlug(ctx, cmd.Slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSportAlreadyExists
	}

	now := time.Now().UTC()
	sp := &Sport{
		Name:        cmd.Name,
		Slug:        cmd.Slug,
		Description: cmd.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = helperdb.Create(ctx, s.db, SportTable, sp, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Sport, error) {
	var sp Sport
	err := helperdb.GetByField(ctx, s.db, SportTable, &Sport{ID: id}, &sp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSportNotFound
		}
		return nil, err
	}
	return &sp, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*Sport, error) {
	var sp Sport
	err := helperdb.GetByField(ctx, s.db, SportTable, &Sport{Slug: slug}, &sp)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (s *service) Search(ctx context.Context, query *SearchSportQuery) (*SearchSportResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name", "slug"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Sport](ctx, s.db, searchHelper.From{Table: SportTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchSportResult{Sports: result.Items, Meta: &m}, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateSportCommand) (*Sport, error) {
	sp, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	cmd.UpdatedAt = time.Now().UTC()
	err = helperdb.Update(ctx, s.db, SportTable, sp.ID, cmd)
	if err != nil {
		return nil, err
	}

	sp.Name = cmd.Name
	sp.Description = cmd.Description
	sp.UpdatedAt = cmd.UpdatedAt

	return sp, nil
}

// sportActiveUpdate is a targeted struct so status changes only touch the relevant columns.
type sportActiveUpdate struct {
	IsActive  bool      `db:"is_active"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Activate(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, SportTable, id, &sportActiveUpdate{
		IsActive:  true,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, SportTable, id, &sportActiveUpdate{
		IsActive:  false,
		UpdatedAt: time.Now().UTC(),
	})
}

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	multiHyphen     = regexp.MustCompile(`-+`)
)

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))

	// Replace anything that isn't a-z or 0-9 with a hyphen.
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")

	// Remove duplicate hyphens.
	slug = multiHyphen.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens.
	slug = strings.Trim(slug, "-")

	return slug
}
