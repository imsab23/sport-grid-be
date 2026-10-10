package tournament

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"sport-grid-be/internal/client"
	"sport-grid-be/internal/sport"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateTournamentCommand) (*Tournament, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error)
	Search(ctx context.Context, query *SearchTournamentQuery) (*SearchTournamentResult, error)
	Update(ctx context.Context, cmd *UpdateTournamentCommand) (*Tournament, error)
	Publish(ctx context.Context, id uuid.UUID) error
	OpenRegistration(ctx context.Context, id uuid.UUID) error
	CloseRegistration(ctx context.Context, id uuid.UUID) error
	Start(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, cmd *CancelTournamentCommand) error
}

type service struct {
	db        db.DB
	sportSvc  sport.Service
	clientSvc client.Service
}

func NewService(
	db db.DB,
	sportSvc sport.Service,
	clientSvc client.Service,
) (Service, error) {
	return &service{
		db:        db,
		sportSvc:  sportSvc,
		clientSvc: clientSvc,
	}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateTournamentCommand) (*Tournament, error) {
	sp, err := s.sportSvc.GetByID(ctx, cmd.SportID)
	if err != nil {
		return nil, err
	}

	if sp.Status != sport.Active {
		return nil, ErrSportNotActive
	}

	_, err = s.clientSvc.GetByID(ctx, cmd.ClientID)
	if err != nil {
		return nil, err
	}

	t := &Tournament{
		ClientID:            cmd.ClientID,
		SportID:             cmd.SportID,
		Name:                cmd.Name,
		Description:         cmd.Description,
		Venue:               cmd.Venue,
		Address:             cmd.Address,
		StartDate:           cmd.StartDate,
		EndDate:             cmd.EndDate,
		RegistrationOpenAt:  cmd.RegistrationOpenAt,
		RegistrationCloseAt: cmd.RegistrationCloseAt,
		RegistrationFee:     cmd.RegistrationFee,
		MaxParticipants:     cmd.MaxParticipants,
		ContactName:         cmd.ContactName,
		ContactEmail:        cmd.ContactEmail,
		ContactPhone:        cmd.ContactPhone,
		Rules:               cmd.Rules,
		Terms:               cmd.Terms,
		PaymentInstructions: cmd.PaymentInstructions,
		Format:              cmd.Format,
		Status:              Draft,
		CreatedBy:           cmd.CreatedBy,
	}

	_, err = helperdb.Create(ctx, s.db, TournamentTable, t, helperdb.CreateOptions{
		ID: helperdb.IDOptions{Mode: helperdb.IDApplication, Force: true},
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error) {
	var result Tournament
	err := helperdb.GetByField(ctx, s.db, TournamentTable, &Tournament{ID: id}, &result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (s *service) Search(ctx context.Context, query *SearchTournamentQuery) (*SearchTournamentResult, error) {
	params := searchHelper.Params{
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Tournament](ctx, s.db,
		searchHelper.From{
			Table: TournamentTable,
		},
		params,
	)
	if err != nil {
		return nil, err
	}

	meta := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))

	return &SearchTournamentResult{
		Tournaments: result.Items,
		Meta:        &meta,
	}, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateTournamentCommand) (*Tournament, error) {
	t, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if t.Status == Completed || t.Status == Cancelled {
		return nil, ErrInvalidTournamentState
	}

	cmd.UpdatedAt = time.Now().UTC()
	err = helperdb.Update(ctx, s.db, TournamentTable, t.ID, cmd)
	if err != nil {
		return nil, err
	}

	t.Name = cmd.Name
	t.Description = cmd.Description
	t.Venue = cmd.Venue
	t.Address = cmd.Address
	t.StartDate = cmd.StartDate
	t.EndDate = cmd.EndDate
	t.RegistrationOpenAt = cmd.RegistrationOpenAt
	t.RegistrationCloseAt = cmd.RegistrationCloseAt
	t.RegistrationFee = cmd.RegistrationFee
	t.MaxParticipants = cmd.MaxParticipants
	t.ContactName = cmd.ContactName
	t.ContactEmail = cmd.ContactEmail
	t.ContactPhone = cmd.ContactPhone
	t.Rules = cmd.Rules
	t.Terms = cmd.Terms
	t.PaymentInstructions = cmd.PaymentInstructions
	t.UpdatedAt = cmd.UpdatedAt

	return t, nil
}

func (s *service) Publish(ctx context.Context, id uuid.UUID) error {
	return s.transition(ctx, id, []Status{Draft}, Published)
}

func (s *service) OpenRegistration(ctx context.Context, id uuid.UUID) error {
	return s.transition(ctx, id, []Status{Published}, RegistrationOpen)
}

func (s *service) CloseRegistration(ctx context.Context, id uuid.UUID) error {
	return s.transition(ctx, id, []Status{RegistrationOpen}, RegistrationClosed)
}

func (s *service) Start(ctx context.Context, id uuid.UUID) error {
	return s.transition(ctx, id, []Status{RegistrationClosed}, Ongoing)
}

func (s *service) Complete(ctx context.Context, id uuid.UUID) error {
	return s.transition(ctx, id, []Status{Ongoing}, Completed)
}

func (s *service) Cancel(ctx context.Context, cmd *CancelTournamentCommand) error {
	t, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	switch t.Status {
	case Draft, Published, RegistrationOpen, RegistrationClosed:
	default:
		return ErrInvalidTournamentState
	}

	return helperdb.Update(ctx, s.db, TournamentTable, cmd.ID, &TournamentCancelUpdate{
		Status:          Cancelled,
		CancelledReason: cmd.Reason,
	})
}

func (s *service) transition(ctx context.Context, id uuid.UUID, from []Status, to Status) error {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	allowed := slices.Contains(from, t.Status)
	if !allowed {
		return ErrInvalidTournamentState
	}

	return helperdb.Update(ctx, s.db, TournamentTable, id, &TournamentStatusUpdate{
		Status: to,
	})
}
