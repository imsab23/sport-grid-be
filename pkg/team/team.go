package team

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/registration"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateTeamCommand) (*Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	Search(ctx context.Context, query *SearchTeamQuery) (*SearchTeamResult, error)
	Disband(ctx context.Context, id uuid.UUID) error

	AddMember(ctx context.Context, cmd *AddTeamMemberCommand) (*TeamMember, error)
	RemoveMember(ctx context.Context, teamID, memberID uuid.UUID) error
	ListMembers(ctx context.Context, teamID uuid.UUID) ([]*TeamMember, error)
}

type service struct {
	db              db.DB
	divisionSvc     division.Service
	registrationSvc registration.Service
}

func NewService(database db.DB, divisionSvc division.Service, registrationSvc registration.Service) (Service, error) {
	return &service{db: database, divisionSvc: divisionSvc, registrationSvc: registrationSvc}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateTeamCommand) (*Team, error) {
	d, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}
	if d.TournamentID != cmd.TournamentID {
		return nil, division.ErrDivisionNotFound
	}
	if d.ParticipantType != division.ParticipantTeam {
		return nil, ErrDivisionNotTeamBased
	}

	reg, err := s.registrationSvc.GetActiveRegistration(ctx, cmd.DivisionID, cmd.CreatorPlayerID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, ErrNotRegisteredForDivision
	}

	var t *Team
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			teamID, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			t = &Team{
				ID:             teamID,
				TournamentID:   cmd.TournamentID,
				DivisionID:     cmd.DivisionID,
				RegistrationID: &reg.ID,
				Name:           cmd.Name,
				Status:         StatusActive,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			_, err = sess.Exec(txCtx, `
				INSERT INTO teams (id, tournament_id, division_id, registration_id, name, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, t.ID, t.TournamentID, t.DivisionID, t.RegistrationID, t.Name, t.Status, t.CreatedAt, t.UpdatedAt)
			if err != nil {
				return err
			}

			memberID, err := uuid.NewV7()
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx, `
				INSERT INTO team_members (id, team_id, player_id, role, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, memberID, t.ID, cmd.CreatorPlayerID, RoleCaptain, now)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	var t Team
	err := helperdb.GetByField(ctx, s.db, TeamTable, &Team{ID: id}, &t)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *service) Search(ctx context.Context, query *SearchTeamQuery) (*SearchTeamResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"name"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Team](ctx, s.db, searchHelper.From{Table: TeamTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchTeamResult{Teams: result.Items, Meta: &m}, nil
}

type teamStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Disband(ctx context.Context, id uuid.UUID) error {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if t.Status == StatusDisbanded {
		return ErrTeamNotActive
	}

	return helperdb.Update(ctx, s.db, TeamTable, id, &teamStatusUpdate{
		Status:    StatusDisbanded,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) AddMember(ctx context.Context, cmd *AddTeamMemberCommand) (*TeamMember, error) {
	t, err := s.GetByID(ctx, cmd.TeamID)
	if err != nil {
		return nil, err
	}
	if t.Status != StatusActive {
		return nil, ErrTeamNotActive
	}

	reg, err := s.registrationSvc.GetActiveRegistration(ctx, t.DivisionID, cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, ErrNotRegisteredForDivision
	}

	var member *TeamMember
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			// Lock the division row so concurrent membership changes across all of its teams are serialized.
			var locked uuid.UUID
			err := sess.Get(txCtx, &locked, `SELECT id FROM tournament_divisions WHERE id = $1 FOR UPDATE`, t.DivisionID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return division.ErrDivisionNotFound
				}
				return err
			}

			var count int
			err = sess.Get(txCtx, &count, `
				SELECT COUNT(*) FROM team_members tm
				JOIN teams tt ON tt.id = tm.team_id
				WHERE tt.division_id = $1 AND tm.player_id = $2
			`, t.DivisionID, cmd.PlayerID)
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrAlreadyOnTeam
			}

			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			member = &TeamMember{ID: id, TeamID: cmd.TeamID, PlayerID: cmd.PlayerID, Role: cmd.Role, CreatedAt: now}

			_, err = sess.Exec(txCtx, `
				INSERT INTO team_members (id, team_id, player_id, role, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, member.ID, member.TeamID, member.PlayerID, member.Role, member.CreatedAt)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (s *service) RemoveMember(ctx context.Context, teamID, memberID uuid.UUID) error {
	_, err := s.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(ctx, `DELETE FROM team_members WHERE id = $1 AND team_id = $2`, memberID, teamID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMemberNotFound
	}

	// Auto-disband a team left with no members to avoid orphaned "ghost" teams.
	var remaining int
	err = s.db.Get(ctx, &remaining, `SELECT COUNT(*) FROM team_members WHERE team_id = $1`, teamID)
	if err != nil {
		return err
	}
	if remaining == 0 {
		return s.Disband(ctx, teamID)
	}

	return nil
}

func (s *service) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*TeamMember, error) {
	members := []*TeamMember{}
	err := s.db.Select(ctx, &members, `SELECT * FROM team_members WHERE team_id = $1 ORDER BY created_at`, teamID)
	if err != nil {
		return nil, err
	}
	return members, nil
}
