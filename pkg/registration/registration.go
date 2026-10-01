package registration

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/tournament"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateRegistrationCommand) (*Registration, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Registration, error)
	Search(ctx context.Context, query *SearchRegistrationQuery) (*SearchRegistrationResult, error)
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
	GetActiveRegistration(ctx context.Context, divisionID, playerID uuid.UUID) (*Registration, error)
	ListConfirmedByDivision(ctx context.Context, divisionID uuid.UUID) ([]*Registration, error)

	CreatePaymentSubmission(ctx context.Context, cmd *CreatePaymentSubmissionCommand) (*PaymentSubmission, error)
	GetPaymentSubmissionByID(ctx context.Context, id uuid.UUID) (*PaymentSubmission, error)
	VerifyPaymentSubmission(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	RejectPaymentSubmission(ctx context.Context, cmd *RejectPaymentSubmissionCommand) error
}

type service struct {
	db            db.DB
	tournamentSvc tournament.Service
	divisionSvc   division.Service
}

func NewService(database db.DB, tournamentSvc tournament.Service, divisionSvc division.Service) (Service, error) {
	return &service{db: database, tournamentSvc: tournamentSvc, divisionSvc: divisionSvc}, nil
}

// GetActiveRegistration finds a non-cancelled registration for the given division/player pair.
func (s *service) GetActiveRegistration(ctx context.Context, divisionID, playerID uuid.UUID) (*Registration, error) {
	var r Registration
	err := s.db.Get(ctx, &r,
		`SELECT * FROM tournament_registrations WHERE division_id = $1 AND player_id = $2 AND status <> 'CANCELLED' LIMIT 1`,
		divisionID, playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListConfirmedByDivision returns confirmed registrations for a division, ordered by registration time.
func (s *service) ListConfirmedByDivision(ctx context.Context, divisionID uuid.UUID) ([]*Registration, error) {
	regs := []*Registration{}
	err := s.db.Select(ctx, &regs,
		`SELECT * FROM tournament_registrations WHERE division_id = $1 AND status = 'CONFIRMED' ORDER BY registered_at ASC`,
		divisionID)
	if err != nil {
		return nil, err
	}
	return regs, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateRegistrationCommand) (*Registration, error) {
	t, err := s.tournamentSvc.GetByID(ctx, cmd.TournamentID)
	if err != nil {
		return nil, err
	}
	if t.Status != tournament.StatusRegistrationOpen {
		return nil, ErrRegistrationNotOpen
	}

	d, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}
	if d.TournamentID != cmd.TournamentID {
		return nil, division.ErrDivisionNotFound
	}
	if d.Status != division.StatusActive {
		return nil, ErrDivisionNotActive
	}

	existing, err := s.GetActiveRegistration(ctx, cmd.DivisionID, cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyRegistered
	}

	var reg *Registration
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			// Lock the division row so concurrent registrations for the last slot are serialized.
			if d.Capacity != nil {
				var locked uuid.UUID
				err := sess.Get(txCtx, &locked, `SELECT id FROM tournament_divisions WHERE id = $1 FOR UPDATE`, cmd.DivisionID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return division.ErrDivisionNotFound
					}
					return err
				}

				var count int
				err = sess.Get(txCtx, &count,
					`SELECT COUNT(*) FROM tournament_registrations WHERE division_id = $1 AND status <> 'CANCELLED'`,
					cmd.DivisionID)
				if err != nil {
					return err
				}

				if count >= *d.Capacity {
					return ErrDivisionFull
				}
			}

			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			reg = &Registration{
				ID:            id,
				TournamentID:  cmd.TournamentID,
				DivisionID:    cmd.DivisionID,
				PlayerID:      cmd.PlayerID,
				Status:        StatusPending,
				PaymentStatus: PaymentStatusNotSubmitted,
				Notes:         cmd.Notes,
				RegisteredAt:  now,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			_, err = sess.Exec(txCtx, `
				INSERT INTO tournament_registrations
					(id, tournament_id, division_id, player_id, status, payment_status, notes, registered_at, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, reg.ID, reg.TournamentID, reg.DivisionID, reg.PlayerID, reg.Status, reg.PaymentStatus, reg.Notes, reg.RegisteredAt, reg.CreatedAt, reg.UpdatedAt)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Registration, error) {
	var r Registration
	err := helperdb.GetByField(ctx, s.db, RegistrationTable, &Registration{ID: id}, &r)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *service) Search(ctx context.Context, query *SearchRegistrationQuery) (*SearchRegistrationResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"id"},
		Filters:     query,
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*Registration](ctx, s.db, searchHelper.From{Table: RegistrationTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchRegistrationResult{Registrations: result.Items, Meta: &m}, nil
}

type registrationApproveUpdate struct {
	Status     Status    `db:"status"`
	ApprovedAt time.Time `db:"approved_at"`
	ApprovedBy uuid.UUID `db:"approved_by"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (s *service) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	r, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if r.Status == StatusCancelled {
		return ErrInvalidRegistrationOp
	}

	return helperdb.Update(ctx, s.db, RegistrationTable, id, &registrationApproveUpdate{
		Status:     StatusConfirmed,
		ApprovedAt: time.Now().UTC(),
		ApprovedBy: approvedBy,
		UpdatedAt:  time.Now().UTC(),
	})
}

type registrationStatusUpdate struct {
	Status    Status    `db:"status"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	r, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if r.Status == StatusCancelled {
		return ErrInvalidRegistrationOp
	}

	return helperdb.Update(ctx, s.db, RegistrationTable, id, &registrationStatusUpdate{
		Status:    StatusCancelled,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *service) resolveExpectedAmount(ctx context.Context, reg *Registration) (float64, error) {
	d, err := s.divisionSvc.GetByID(ctx, reg.DivisionID)
	if err != nil {
		return 0, err
	}
	if d.RegistrationFee != nil {
		return *d.RegistrationFee, nil
	}

	t, err := s.tournamentSvc.GetByID(ctx, reg.TournamentID)
	if err != nil {
		return 0, err
	}
	return t.RegistrationFee, nil
}

func (s *service) CreatePaymentSubmission(ctx context.Context, cmd *CreatePaymentSubmissionCommand) (*PaymentSubmission, error) {
	reg, err := s.GetByID(ctx, cmd.RegistrationID)
	if err != nil {
		return nil, err
	}
	if reg.Status == StatusCancelled {
		return nil, ErrInvalidRegistrationOp
	}

	amountExpected, err := s.resolveExpectedAmount(ctx, reg)
	if err != nil {
		return nil, err
	}

	amountStatus := AmountExact
	switch {
	case cmd.AmountSubmitted < amountExpected:
		amountStatus = AmountUnderpaid
	case cmd.AmountSubmitted > amountExpected:
		amountStatus = AmountOverpaid
	}

	var ps *PaymentSubmission
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			ps = &PaymentSubmission{
				ID:              id,
				RegistrationID:  cmd.RegistrationID,
				AmountExpected:  amountExpected,
				AmountSubmitted: cmd.AmountSubmitted,
				PaymentMethod:   cmd.PaymentMethod,
				ReferenceNumber: cmd.ReferenceNumber,
				ProofReference:  cmd.ProofReference,
				Status:          PaymentSubmissionSubmitted,
				AmountStatus:    amountStatus,
				SubmittedAt:     now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			_, err = sess.Exec(txCtx, `
				INSERT INTO payment_submissions
					(id, registration_id, amount_expected, amount_submitted, payment_method, reference_number, proof_reference, status, amount_status, submitted_at, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`, ps.ID, ps.RegistrationID, ps.AmountExpected, ps.AmountSubmitted, ps.PaymentMethod, ps.ReferenceNumber, ps.ProofReference, ps.Status, ps.AmountStatus, ps.SubmittedAt, ps.CreatedAt, ps.UpdatedAt)
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx,
				`UPDATE tournament_registrations SET payment_status = $1, updated_at = $2 WHERE id = $3`,
				PaymentStatusSubmitted, now, reg.ID)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *service) GetPaymentSubmissionByID(ctx context.Context, id uuid.UUID) (*PaymentSubmission, error) {
	var ps PaymentSubmission
	err := helperdb.GetByField(ctx, s.db, PaymentSubmissionTable, &PaymentSubmission{ID: id}, &ps)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &ps, nil
}

func (s *service) VerifyPaymentSubmission(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	ps, err := s.GetPaymentSubmissionByID(ctx, id)
	if err != nil {
		return err
	}
	if ps.Status != PaymentSubmissionSubmitted {
		return ErrInvalidPaymentOp
	}

	now := time.Now().UTC()
	return s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			_, err := sess.Exec(txCtx,
				`UPDATE payment_submissions SET status = $1, verified_at = $2, verified_by = $3, updated_at = $4 WHERE id = $5`,
				PaymentSubmissionVerified, now, verifiedBy, now, ps.ID)
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx,
				`UPDATE tournament_registrations SET payment_status = $1, status = $2, approved_at = $3, approved_by = $4, updated_at = $5 WHERE id = $6`,
				PaymentStatusVerified, StatusConfirmed, now, verifiedBy, now, ps.RegistrationID)
			return err
		})
	})
}

func (s *service) RejectPaymentSubmission(ctx context.Context, cmd *RejectPaymentSubmissionCommand) error {
	ps, err := s.GetPaymentSubmissionByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if ps.Status != PaymentSubmissionSubmitted {
		return ErrInvalidPaymentOp
	}

	now := time.Now().UTC()
	return s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			_, err := sess.Exec(txCtx,
				`UPDATE payment_submissions SET status = $1, rejection_reason = $2, updated_at = $3 WHERE id = $4`,
				PaymentSubmissionRejected, cmd.Reason, now, ps.ID)
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx,
				`UPDATE tournament_registrations SET payment_status = $1, updated_at = $2 WHERE id = $3`,
				PaymentStatusRejected, now, ps.RegistrationID)
			return err
		})
	})
}
