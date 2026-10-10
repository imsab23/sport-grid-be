package service

import (
	"context"
	"database/sql"
	"errors"
	"sport-grid-be/internal/division"
	"sport-grid-be/internal/registration"
	"sport-grid-be/internal/tournament"
	"time"

	"github.com/google/uuid"
	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
)

type service struct {
	db            db.DB
	tournamentSvc tournament.Service
	divisionSvc   division.Service
}

func NewService(
	db db.DB,
	tournamentSvc tournament.Service,
	divisionSvc division.Service,
) (registration.Service, error) {
	return &service{
		db:            db,
		tournamentSvc: tournamentSvc,
		divisionSvc:   divisionSvc,
	}, nil
}

func (s *service) GetActiveRegistration(ctx context.Context, divisionID, playerID uuid.UUID) (*registration.Registration, error) {
	var result registration.Registration

	rawSQL := `
		SELECT 
			* 
		FROM tournament_registrations 
		WHERE 
			division_id = $1 AND 
			player_id = $2 AND 
			status <> 'CANCELLED' 
		LIMIT 1
	`

	err := s.db.Get(ctx, &result, rawSQL, divisionID, playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *service) ListConfirmedByDivision(ctx context.Context, divisionID uuid.UUID) ([]*registration.Registration, error) {
	result := []*registration.Registration{}

	rawSQL := `
		SELECT * 
		FROM tournament_registrations 
		WHERE 
			division_id = $1 AND 
			status = 'CONFIRMED' 
		ORDER BY 
			registered_at ASC
	`

	err := s.db.Select(ctx, &result, rawSQL, divisionID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) Create(ctx context.Context, cmd *registration.CreateRegistrationCommand) (*registration.Registration, error) {
	t, err := s.tournamentSvc.GetByID(ctx, cmd.TournamentID)
	if err != nil {
		return nil, err
	}

	if t.Status != tournament.RegistrationOpen {
		return nil, registration.ErrRegistrationNotOpen
	}

	d, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}

	if d.TournamentID != cmd.TournamentID {
		return nil, division.ErrDivisionNotFound
	}

	if d.Status != division.StatusActive {
		return nil, registration.ErrDivisionNotActive
	}

	existing, err := s.GetActiveRegistration(ctx, cmd.DivisionID, cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, registration.ErrAlreadyRegistered
	}

	var result *registration.Registration
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
					return registration.ErrDivisionFull
				}
			}

			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			result = &registration.Registration{
				ID:            id,
				TournamentID:  cmd.TournamentID,
				DivisionID:    cmd.DivisionID,
				PlayerID:      cmd.PlayerID,
				Status:        registration.Pending,
				PaymentStatus: registration.NotSubmitted,
				Notes:         cmd.Notes,
				RegisteredAt:  now,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			_, err = helperdb.Create(txCtx, s.db, registration.RegistrationTable, result)
			if err != nil {
				return err
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*registration.Registration, error) {
	var r registration.Registration
	err := helperdb.GetByField(ctx, s.db, registration.RegistrationTable, &registration.Registration{ID: id}, &r)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, registration.ErrRegistrationNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *service) Search(ctx context.Context, query *registration.SearchRegistrationQuery) (*registration.SearchRegistrationResult, error) {
	params := searchHelper.Params{
		Filters:     query,
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*registration.Registration](ctx, s.db, searchHelper.From{Table: registration.RegistrationTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &registration.SearchRegistrationResult{Registrations: result.Items, Meta: &m}, nil
}

func (s *service) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	result, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if result.Status == registration.Cancelled {
		return registration.ErrInvalidRegistrationOp
	}

	return helperdb.Update(ctx, s.db, registration.RegistrationTable, id, &registration.RegistrationApproveUpdate{
		Status:     registration.Confirmed,
		ApprovedAt: time.Now().UTC(),
		ApprovedBy: approvedBy,
	})
}

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	r, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if r.Status == registration.Cancelled {
		return registration.ErrInvalidRegistrationOp
	}

	return helperdb.Update(ctx, s.db, registration.RegistrationTable, id, &registration.RegistrationStatusUpdate{
		Status: registration.Cancelled,
	})
}

func (s *service) resolveExpectedAmount(ctx context.Context, reg *registration.Registration) (float64, error) {
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

func (s *service) CreatePaymentSubmission(ctx context.Context, cmd *registration.CreatePaymentSubmissionCommand) (*registration.PaymentSubmission, error) {
	reg, err := s.GetByID(ctx, cmd.RegistrationID)
	if err != nil {
		return nil, err
	}
	if reg.Status == registration.Cancelled {
		return nil, registration.ErrInvalidRegistrationOp
	}

	amountExpected, err := s.resolveExpectedAmount(ctx, reg)
	if err != nil {
		return nil, err
	}

	amountStatus := registration.AmountExact
	switch {
	case cmd.AmountSubmitted < amountExpected:
		amountStatus = registration.AmountUnderpaid
	case cmd.AmountSubmitted > amountExpected:
		amountStatus = registration.AmountOverpaid
	}

	var ps *registration.PaymentSubmission
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			ps = &registration.PaymentSubmission{
				ID:              id,
				RegistrationID:  cmd.RegistrationID,
				AmountExpected:  amountExpected,
				AmountSubmitted: cmd.AmountSubmitted,
				PaymentMethod:   cmd.PaymentMethod,
				ReferenceNumber: cmd.ReferenceNumber,
				ProofReference:  cmd.ProofReference,
				Status:          registration.PaymentSubmissionSubmitted,
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
				registration.Submitted, now, reg.ID)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *service) GetPaymentSubmissionByID(ctx context.Context, id uuid.UUID) (*registration.PaymentSubmission, error) {
	var ps registration.PaymentSubmission
	err := helperdb.GetByField(ctx, s.db, registration.PaymentSubmissionTable, &registration.PaymentSubmission{ID: id}, &ps)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, registration.ErrPaymentNotFound
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
	if ps.Status != registration.PaymentSubmissionSubmitted {
		return registration.ErrInvalidPaymentOp
	}

	now := time.Now().UTC()
	return s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			_, err := sess.Exec(txCtx,
				`UPDATE payment_submissions SET status = $1, verified_at = $2, verified_by = $3, updated_at = $4 WHERE id = $5`,
				registration.PaymentSubmissionVerified, now, verifiedBy, now, ps.ID)
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx,
				`UPDATE tournament_registrations SET payment_status = $1, status = $2, approved_at = $3, approved_by = $4, updated_at = $5 WHERE id = $6`,
				registration.Verified, registration.Confirmed, now, verifiedBy, now, ps.RegistrationID)
			return err
		})
	})
}

func (s *service) RejectPaymentSubmission(ctx context.Context, cmd *registration.RejectPaymentSubmissionCommand) error {
	ps, err := s.GetPaymentSubmissionByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if ps.Status != registration.PaymentSubmissionSubmitted {
		return registration.ErrInvalidPaymentOp
	}

	now := time.Now().UTC()
	return s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			_, err := sess.Exec(txCtx,
				`UPDATE payment_submissions SET status = $1, rejection_reason = $2, updated_at = $3 WHERE id = $4`,
				registration.PaymentSubmissionRejected, cmd.Reason, now, ps.ID)
			if err != nil {
				return err
			}

			_, err = sess.Exec(txCtx,
				`UPDATE tournament_registrations SET payment_status = $1, updated_at = $2 WHERE id = $3`,
				registration.Rejected, now, ps.RegistrationID)
			return err
		})
	})
}
