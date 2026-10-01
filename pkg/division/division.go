package division

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
	Create(ctx context.Context, cmd *CreateDivisionCommand) (*Division, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Division, error)
	Search(ctx context.Context, query *SearchDivisionQuery) (*SearchDivisionResult, error)
	Update(ctx context.Context, cmd *UpdateDivisionCommand) (*Division, error)
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
}

type service struct {
	db            db.DB
	tournamentSvc tournament.Service
}

func NewService(database db.DB, tournamentSvc tournament.Service) (Service, error) {
	return &service{db: database, tournamentSvc: tournamentSvc}, nil
}

// requireEditableTournament fails if the parent tournament is in a terminal state.
func (s *service) requireEditableTournament(ctx context.Context, tournamentID uuid.UUID) error {
	t, err := s.tournamentSvc.GetByID(ctx, tournamentID)
	if err != nil {
		return err
	}
	if t.Status == tournament.StatusCompleted || t.Status == tournament.StatusCancelled {
		return ErrTournamentNotEditable
	}
	return nil
}

func (s *service) Create(ctx context.Context, cmd *CreateDivisionCommand) (*Division, error) {
	err := s.requireEditableTournament(ctx, cmd.TournamentID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	d := &Division{
		TournamentID:    cmd.TournamentID,
		Name:            cmd.Name,
		Gender:          cmd.Gender,
		MinAge:          cmd.MinAge,
		MaxAge:          cmd.MaxAge,
		SkillLevel:      cmd.SkillLevel,
		ParticipantType: cmd.ParticipantType,
		Capacity:        cmd.Capacity,
		RegistrationFee: cmd.RegistrationFee,
		Format:          cmd.Format,
		Status:          StatusActive,
		BestOf:          cmd.BestOf,
		TargetScore:     cmd.TargetScore,
		WinBy:           cmd.WinBy,
		MaxScore:        cmd.MaxScore,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err = helperdb.Create(ctx, s.db, DivisionTable, d, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Division, error) {
	var d Division
	err := helperdb.GetByField(ctx, s.db, DivisionTable, &Division{ID: id}, &d)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDivisionNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (s *service) Search(ctx context.Context, query *SearchDivisionQuery) (*SearchDivisionResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Division](ctx, s.db, searchHelper.From{Table: DivisionTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchDivisionResult{Divisions: result.Items, Meta: &m}, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateDivisionCommand) (*Division, error) {
	d, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	err = s.requireEditableTournament(ctx, d.TournamentID)
	if err != nil {
		return nil, err
	}

	cmd.UpdatedAt = time.Now().UTC()
	err = helperdb.Update(ctx, s.db, DivisionTable, d.ID, cmd)
	if err != nil {
		return nil, err
	}

	d.Name = cmd.Name
	d.Gender = cmd.Gender
	d.MinAge = cmd.MinAge
	d.MaxAge = cmd.MaxAge
	d.SkillLevel = cmd.SkillLevel
	d.ParticipantType = cmd.ParticipantType
	d.Capacity = cmd.Capacity
	d.RegistrationFee = cmd.RegistrationFee
	d.Format = cmd.Format
	d.BestOf = cmd.BestOf
	d.TargetScore = cmd.TargetScore
	d.WinBy = cmd.WinBy
	d.MaxScore = cmd.MaxScore
	d.UpdatedAt = cmd.UpdatedAt

	return d, nil
}

// divisionStatusUpdate is a targeted struct so status changes only touch the relevant columns.
type divisionStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Activate(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, DivisionTable, id, &divisionStatusUpdate{
		Status:    StatusActive,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return helperdb.Update(ctx, s.db, DivisionTable, id, &divisionStatusUpdate{
		Status:    StatusInactive,
		UpdatedAt: time.Now().UTC(),
	})
}
