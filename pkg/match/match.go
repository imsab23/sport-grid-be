package match

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/pkg/bracket"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	GenerateMatches(ctx context.Context, bracketID uuid.UUID) ([]*Match, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Match, error)
	GetParticipants(ctx context.Context, matchID uuid.UUID) ([]*MatchParticipant, error)
	Search(ctx context.Context, query *SearchMatchQuery) (*SearchMatchResult, error)
	Schedule(ctx context.Context, cmd *ScheduleMatchCommand) (*Match, error)
	Start(ctx context.Context, matchID uuid.UUID) error
	RecordGame(ctx context.Context, cmd *RecordGameCommand) (*MatchGame, error)
	GetGames(ctx context.Context, matchID uuid.UUID) ([]*MatchGame, error)
	Finalize(ctx context.Context, cmd *FinalizeMatchCommand) error
	Cancel(ctx context.Context, cmd *CancelMatchCommand) error
}

type service struct {
	db         db.DB
	bracketSvc bracket.Service
}

func NewService(database db.DB, bracketSvc bracket.Service) (Service, error) {
	return &service{db: database, bracketSvc: bracketSvc}, nil
}

// GenerateMatches scans the bracket tree for node-pairs whose children both have a resolved
// participant and no match yet, creating a Match for each newly-ready pairing. It is idempotent
// and incremental — call it again after each Finalize to materialize the next round's matches.
func (s *service) GenerateMatches(ctx context.Context, bracketID uuid.UUID) ([]*Match, error) {
	b, err := s.bracketSvc.GetByID(ctx, bracketID)
	if err != nil {
		return nil, err
	}

	nodes, err := s.bracketSvc.GetNodes(ctx, bracketID)
	if err != nil {
		return nil, err
	}

	childrenByParent := map[uuid.UUID][]*bracket.BracketNode{}
	for _, n := range nodes {
		if n.NextNodeID != nil {
			childrenByParent[*n.NextNodeID] = append(childrenByParent[*n.NextNodeID], n)
		}
	}

	existing := []*Match{}
	err = s.db.Select(ctx, &existing, `SELECT * FROM matches WHERE bracket_id = $1`, bracketID)
	if err != nil {
		return nil, err
	}
	hasMatch := make(map[uuid.UUID]bool, len(existing))
	for _, m := range existing {
		if m.BracketNodeID != nil {
			hasMatch[*m.BracketNodeID] = true
		}
	}

	var created []*Match
	now := time.Now().UTC()

	for _, parent := range nodes {
		if parent.Round == 1 || hasMatch[parent.ID] {
			continue
		}

		children := childrenByParent[parent.ID]
		if len(children) != 2 {
			continue
		}
		c1, c2 := children[0], children[1]
		if c1.RegistrationID == nil || c2.RegistrationID == nil {
			continue
		}

		matchID, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		m := &Match{
			ID:            matchID,
			TournamentID:  b.TournamentID,
			DivisionID:    b.DivisionID,
			BracketID:     &b.ID,
			BracketNodeID: &parent.ID,
			Round:         parent.Round - 1,
			MatchNumber:   parent.Position + 1,
			Status:        StatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
			return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
				_, err := sess.Exec(txCtx, `
					INSERT INTO matches (id, tournament_id, division_id, bracket_id, bracket_node_id, round, match_number, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				`, m.ID, m.TournamentID, m.DivisionID, m.BracketID, m.BracketNodeID, m.Round, m.MatchNumber, m.Status, m.CreatedAt, m.UpdatedAt)
				if err != nil {
					return err
				}

				for side, child := range map[int]*bracket.BracketNode{1: c1, 2: c2} {
					pid, err := uuid.NewV7()
					if err != nil {
						return err
					}
					_, err = sess.Exec(txCtx, `
						INSERT INTO match_participants (id, match_id, registration_id, side, is_winner, created_at)
						VALUES ($1, $2, $3, $4, $5, $6)
					`, pid, m.ID, child.RegistrationID, side, false, now)
					if err != nil {
						return err
					}
				}
				return nil
			})
		})
		if err != nil {
			return nil, err
		}

		created = append(created, m)
	}

	return created, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Match, error) {
	var m Match
	err := helperdb.GetByField(ctx, s.db, MatchTable, &Match{ID: id}, &m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (s *service) GetParticipants(ctx context.Context, matchID uuid.UUID) ([]*MatchParticipant, error) {
	participants := []*MatchParticipant{}
	err := s.db.Select(ctx, &participants,
		`SELECT * FROM match_participants WHERE match_id = $1 ORDER BY side ASC`, matchID)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (s *service) Search(ctx context.Context, query *SearchMatchQuery) (*SearchMatchResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Match](ctx, s.db, searchHelper.From{Table: MatchTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchMatchResult{Matches: result.Items, Meta: &m}, nil
}

type matchScheduleUpdate struct {
	CourtID     *uuid.UUID `db:"court_id"`
	ScheduledAt *time.Time `db:"scheduled_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func (s *service) Schedule(ctx context.Context, cmd *ScheduleMatchCommand) (*Match, error) {
	m, err := s.GetByID(ctx, cmd.MatchID)
	if err != nil {
		return nil, err
	}
	if m.Status == StatusCompleted || m.Status == StatusCancelled {
		return nil, ErrMatchNotEditable
	}

	update := &matchScheduleUpdate{CourtID: cmd.CourtID, ScheduledAt: cmd.ScheduledAt, UpdatedAt: time.Now().UTC()}
	err = helperdb.Update(ctx, s.db, MatchTable, m.ID, update)
	if err != nil {
		return nil, err
	}

	m.CourtID = cmd.CourtID
	m.ScheduledAt = cmd.ScheduledAt
	m.UpdatedAt = update.UpdatedAt
	return m, nil
}

type matchStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Start(ctx context.Context, matchID uuid.UUID) error {
	m, err := s.GetByID(ctx, matchID)
	if err != nil {
		return err
	}
	if m.Status != StatusScheduled && m.Status != StatusReady {
		return ErrMatchNotEditable
	}

	return helperdb.Update(ctx, s.db, MatchTable, matchID, &matchStatusUpdate{
		Status:    StatusInProgress,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) GetGames(ctx context.Context, matchID uuid.UUID) ([]*MatchGame, error) {
	games := []*MatchGame{}
	err := s.db.Select(ctx, &games,
		`SELECT * FROM match_games WHERE match_id = $1 ORDER BY game_number ASC`, matchID)
	if err != nil {
		return nil, err
	}
	return games, nil
}

// RecordGame creates or updates the score for a given game number (idempotent score entry).
func (s *service) RecordGame(ctx context.Context, cmd *RecordGameCommand) (*MatchGame, error) {
	m, err := s.GetByID(ctx, cmd.MatchID)
	if err != nil {
		return nil, err
	}
	if m.Status == StatusCompleted || m.Status == StatusCancelled {
		return nil, ErrMatchNotEditable
	}

	now := time.Now().UTC()

	var existing MatchGame
	err = s.db.Get(ctx, &existing,
		`SELECT * FROM match_games WHERE match_id = $1 AND game_number = $2`, cmd.MatchID, cmd.GameNumber)
	if errors.Is(err, sql.ErrNoRows) {
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		g := &MatchGame{
			ID:         id,
			MatchID:    cmd.MatchID,
			GameNumber: cmd.GameNumber,
			ScoreSide1: cmd.ScoreSide1,
			ScoreSide2: cmd.ScoreSide2,
			IsTiebreak: cmd.IsTiebreak,
			Status:     GameCompleted,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		_, err = s.db.Exec(ctx, `
			INSERT INTO match_games (id, match_id, game_number, score_side1, score_side2, is_tiebreak, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, g.ID, g.MatchID, g.GameNumber, g.ScoreSide1, g.ScoreSide2, g.IsTiebreak, g.Status, g.CreatedAt, g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return g, nil
	}
	if err != nil {
		return nil, err
	}

	existing.ScoreSide1 = cmd.ScoreSide1
	existing.ScoreSide2 = cmd.ScoreSide2
	existing.IsTiebreak = cmd.IsTiebreak
	existing.Status = GameCompleted
	existing.UpdatedAt = now

	_, err = s.db.Exec(ctx, `
		UPDATE match_games SET score_side1 = $1, score_side2 = $2, is_tiebreak = $3, status = $4, updated_at = $5 WHERE id = $6
	`, existing.ScoreSide1, existing.ScoreSide2, existing.IsTiebreak, existing.Status, existing.UpdatedAt, existing.ID)
	if err != nil {
		return nil, err
	}

	return &existing, nil
}

// Finalize records the winner (staff-declared, scores are informational only — no
// sport-specific auto-win-detection is implemented) and propagates the winner into the bracket.
func (s *service) Finalize(ctx context.Context, cmd *FinalizeMatchCommand) error {
	m, err := s.GetByID(ctx, cmd.MatchID)
	if err != nil {
		return err
	}
	if m.Status == StatusCancelled {
		return ErrMatchNotEditable
	}

	participants, err := s.GetParticipants(ctx, m.ID)
	if err != nil {
		return err
	}

	var winner *MatchParticipant
	for _, p := range participants {
		if p.Side == cmd.WinnerSide {
			winner = p
			break
		}
	}
	if winner == nil {
		return ErrInvalidWinnerSide
	}

	now := time.Now().UTC()

	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			// Atomic guard against a concurrent double-finalize of the same match.
			result, err := sess.Exec(txCtx, `
				UPDATE matches
				SET status = $1, result_type = $2, winner_registration_id = $3, notes = $4, finalized_by = $5, finalized_at = $6, updated_at = $7
				WHERE id = $8 AND status NOT IN ('COMPLETED', 'CANCELLED')
			`, StatusCompleted, cmd.ResultType, winner.RegistrationID, cmd.Notes, cmd.FinalizedBy, now, now, m.ID)
			if err != nil {
				return err
			}
			rows, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return ErrMatchAlreadyFinal
			}

			_, err = sess.Exec(txCtx,
				`UPDATE match_participants SET is_winner = true WHERE match_id = $1 AND side = $2`,
				m.ID, cmd.WinnerSide)
			return err
		})
	})
	if err != nil {
		return err
	}

	if m.BracketNodeID != nil && winner.RegistrationID != nil {
		return s.bracketSvc.AdvanceWinner(ctx, *m.BracketNodeID, *winner.RegistrationID)
	}

	return nil
}

func (s *service) Cancel(ctx context.Context, cmd *CancelMatchCommand) error {
	m, err := s.GetByID(ctx, cmd.MatchID)
	if err != nil {
		return err
	}
	if m.Status == StatusCompleted || m.Status == StatusCancelled {
		return ErrMatchNotEditable
	}

	return helperdb.Update(ctx, s.db, MatchTable, cmd.MatchID, &struct {
		Status    Status    `db:"status"`
		Notes     *string   `db:"notes"`
		UpdatedAt time.Time `db:"updated_at"`
	}{
		Status:    StatusCancelled,
		Notes:     cmd.Reason,
		UpdatedAt: time.Now().UTC(),
	})
}
