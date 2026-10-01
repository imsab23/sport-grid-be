package bracket

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/seeding"
	"sport-grid-be/pkg/tournament"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
)

type Service interface {
	Generate(ctx context.Context, cmd *GenerateBracketCommand) (*Bracket, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Bracket, error)
	GetNodes(ctx context.Context, bracketID uuid.UUID) ([]*BracketNode, error)
	Reset(ctx context.Context, cmd *ResetBracketCommand) (*Bracket, error)
	AdvanceWinner(ctx context.Context, nodeID uuid.UUID, registrationID uuid.UUID) error
}

type service struct {
	db            db.DB
	divisionSvc   division.Service
	tournamentSvc tournament.Service
	seedingSvc    seeding.Service
}

func NewService(database db.DB, divisionSvc division.Service, tournamentSvc tournament.Service, seedingSvc seeding.Service) (Service, error) {
	return &service{db: database, divisionSvc: divisionSvc, tournamentSvc: tournamentSvc, seedingSvc: seedingSvc}, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Bracket, error) {
	var b Bracket
	err := helperdb.GetByField(ctx, s.db, BracketTable, &Bracket{ID: id}, &b)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBracketNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (s *service) GetNodes(ctx context.Context, bracketID uuid.UUID) ([]*BracketNode, error) {
	nodes := []*BracketNode{}
	err := s.db.Select(ctx, &nodes,
		`SELECT * FROM bracket_nodes WHERE bracket_id = $1 ORDER BY round ASC, position ASC`, bracketID)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// AdvanceWinner records a match winner onto the bracket node they advanced to.
func (s *service) AdvanceWinner(ctx context.Context, nodeID uuid.UUID, registrationID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE bracket_nodes SET registration_id = $1, updated_at = $2 WHERE id = $3`,
		registrationID, time.Now().UTC(), nodeID)
	return err
}

// nextPowerOfTwo returns the smallest power of 2 >= n.
func nextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

// seedOrder returns the standard single-elimination bracket placement order for `size` slots,
// e.g. size=8 -> [1,8,4,5,2,7,3,6]. Seeds this far apart meet as late as possible.
func seedOrder(size int) []int {
	order := []int{1}
	for len(order) < size {
		next := make([]int, 0, len(order)*2)
		sum := len(order)*2 + 1
		for _, v := range order {
			next = append(next, v, sum-v)
		}
		order = next
	}
	return order
}

func (s *service) Generate(ctx context.Context, cmd *GenerateBracketCommand) (*Bracket, error) {
	d, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}
	// Scope: team-based bracket generation (resolving a team's representative registration
	// to a team_id) is not yet implemented — deferred until Team gains broader exposure.
	if d.ParticipantType != division.ParticipantIndividual {
		return nil, ErrTeamDivisionNotSupported
	}

	t, err := s.tournamentSvc.GetByID(ctx, d.TournamentID)
	if err != nil {
		return nil, err
	}
	if t.Format != tournament.FormatSingleElimination {
		return nil, ErrUnsupportedBracketFormat
	}

	seeds, err := s.seedingSvc.GetByDivision(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}
	if len(seeds) == 0 {
		return nil, ErrNoSeedingsFound
	}
	if len(seeds) < 2 {
		return nil, ErrInsufficientSeeds
	}

	n := len(seeds)
	size := nextPowerOfTwo(n)
	order := seedOrder(size)

	// layerSizes[0] = size (round 1 leaves) ... layerSizes[last] = 1 (champion slot).
	layerSizes := []int{}
	for sz := size; sz >= 1; sz /= 2 {
		layerSizes = append(layerSizes, sz)
	}

	layerNodeIDs := make([][]uuid.UUID, len(layerSizes))
	for li, sz := range layerSizes {
		ids := make([]uuid.UUID, sz)
		for i := range ids {
			id, err := uuid.NewV7()
			if err != nil {
				return nil, err
			}
			ids[i] = id
		}
		layerNodeIDs[li] = ids
	}

	now := time.Now().UTC()
	bracketID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	b := &Bracket{
		ID:           bracketID,
		TournamentID: d.TournamentID,
		DivisionID:   cmd.DivisionID,
		Format:       string(tournament.FormatSingleElimination),
		Status:       StatusActive,
		GeneratedBy:  cmd.GeneratedBy,
		GeneratedAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	nodes := make([]*BracketNode, 0, 2*size-1)
	leafRegByPos := make(map[int]*uuid.UUID, size)

	for li, sz := range layerSizes {
		for pos := 0; pos < sz; pos++ {
			var nextID *uuid.UUID
			if li+1 < len(layerSizes) {
				id := layerNodeIDs[li+1][pos/2]
				nextID = &id
			}

			node := &BracketNode{
				ID:         layerNodeIDs[li][pos],
				BracketID:  bracketID,
				Round:      li + 1,
				Position:   pos,
				NextNodeID: nextID,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if li == 0 {
				seedNum := order[pos]
				if seedNum <= n {
					regID := seeds[seedNum-1].RegistrationID
					node.RegistrationID = &regID
					leafRegByPos[pos] = &regID
				} else {
					node.IsBye = true
				}
			}

			nodes = append(nodes, node)
		}
	}

	// Resolve round-1 byes by auto-advancing the sole real participant into round 2.
	// The seedOrder construction guarantees a bye never pairs against another bye
	// (size is always < 2n for the smallest power-of-two >= n), so this single pass suffices.
	round2 := nodes[size:] // flat slice offset: round-1 occupies indices [0, size)
	for pos := 0; pos < size; pos += 2 {
		a, bNode := nodes[pos], nodes[pos+1]
		switch {
		case a.IsBye && !bNode.IsBye:
			round2[pos/2].RegistrationID = leafRegByPos[pos+1]
		case bNode.IsBye && !a.IsBye:
			round2[pos/2].RegistrationID = leafRegByPos[pos]
		}
	}

	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			_, err := sess.Exec(txCtx, `
				INSERT INTO brackets (id, tournament_id, division_id, format, status, generated_by, generated_at, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, b.ID, b.TournamentID, b.DivisionID, b.Format, b.Status, b.GeneratedBy, b.GeneratedAt, b.CreatedAt, b.UpdatedAt)
			if err != nil {
				return err
			}

			for _, node := range nodes {
				_, err = sess.Exec(txCtx, `
					INSERT INTO bracket_nodes (id, bracket_id, round, position, registration_id, team_id, is_bye, next_node_id, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				`, node.ID, node.BracketID, node.Round, node.Position, node.RegistrationID, node.TeamID, node.IsBye, node.NextNodeID, node.CreatedAt, node.UpdatedAt)
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

	return b, nil
}

type bracketResetUpdate struct {
	Status      Status    `db:"status"`
	ResetReason string    `db:"reset_reason"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Reset marks the current bracket as RESET and generates a fresh one in its place.
func (s *service) Reset(ctx context.Context, cmd *ResetBracketCommand) (*Bracket, error) {
	old, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if old.Status == StatusReset {
		return nil, ErrBracketNotFound
	}

	err = helperdb.Update(ctx, s.db, BracketTable, old.ID, &bracketResetUpdate{
		Status:      StatusReset,
		ResetReason: cmd.Reason,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	return s.Generate(ctx, &GenerateBracketCommand{DivisionID: old.DivisionID, GeneratedBy: cmd.GeneratedBy})
}
