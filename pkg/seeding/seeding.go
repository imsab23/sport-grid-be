package seeding

import (
	"context"
	"math/rand"
	"time"

	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/registration"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
)

type Service interface {
	AssignSeed(ctx context.Context, cmd *AssignSeedCommand) (*Seeding, error)
	GenerateSeeding(ctx context.Context, cmd *GenerateSeedingCommand) ([]*Seeding, error)
	GetByDivision(ctx context.Context, divisionID uuid.UUID) ([]*Seeding, error)
}

type service struct {
	db              db.DB
	divisionSvc     division.Service
	registrationSvc registration.Service
}

func NewService(database db.DB, divisionSvc division.Service, registrationSvc registration.Service) (Service, error) {
	return &service{db: database, divisionSvc: divisionSvc, registrationSvc: registrationSvc}, nil
}

func (s *service) GetByDivision(ctx context.Context, divisionID uuid.UUID) ([]*Seeding, error) {
	seeds := []*Seeding{}
	err := s.db.Select(ctx, &seeds,
		`SELECT * FROM seedings WHERE division_id = $1 ORDER BY seed_number ASC`, divisionID)
	if err != nil {
		return nil, err
	}
	return seeds, nil
}

func (s *service) AssignSeed(ctx context.Context, cmd *AssignSeedCommand) (*Seeding, error) {
	_, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}

	reg, err := s.registrationSvc.GetByID(ctx, cmd.RegistrationID)
	if err != nil {
		return nil, err
	}
	if reg.Status != registration.StatusConfirmed {
		return nil, ErrRegistrationNotConfirmed
	}

	var sd *Seeding
	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			var taken int
			err := sess.Get(txCtx, &taken,
				`SELECT COUNT(*) FROM seedings WHERE division_id = $1 AND seed_number = $2 AND registration_id <> $3`,
				cmd.DivisionID, cmd.SeedNumber, cmd.RegistrationID)
			if err != nil {
				return err
			}
			if taken > 0 {
				return ErrSeedNumberTaken
			}

			// Replace any existing seed for this registration (re-assignment).
			_, err = sess.Exec(txCtx,
				`DELETE FROM seedings WHERE division_id = $1 AND registration_id = $2`,
				cmd.DivisionID, cmd.RegistrationID)
			if err != nil {
				return err
			}

			id, err := uuid.NewV7()
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			sd = &Seeding{
				ID:             id,
				DivisionID:     cmd.DivisionID,
				RegistrationID: cmd.RegistrationID,
				SeedNumber:     cmd.SeedNumber,
				Method:         MethodManual,
				SeededBy:       cmd.SeededBy,
				SeededAt:       now,
				CreatedAt:      now,
			}

			_, err = sess.Exec(txCtx, `
				INSERT INTO seedings (id, division_id, registration_id, seed_number, method, seeded_by, seeded_at, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, sd.ID, sd.DivisionID, sd.RegistrationID, sd.SeedNumber, sd.Method, sd.SeededBy, sd.SeededAt, sd.CreatedAt)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

	return sd, nil
}

func (s *service) GenerateSeeding(ctx context.Context, cmd *GenerateSeedingCommand) ([]*Seeding, error) {
	_, err := s.divisionSvc.GetByID(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}

	regs, err := s.registrationSvc.ListConfirmedByDivision(ctx, cmd.DivisionID)
	if err != nil {
		return nil, err
	}
	if len(regs) < 2 {
		return nil, ErrInsufficientRegistrations
	}

	order := make([]int, len(regs))
	switch cmd.Method {
	case MethodRegistrationOrder:
		// regs is already ordered by registration time ascending.
		for i := range order {
			order[i] = i
		}
	case MethodRandom:
		order = rand.Perm(len(regs))
	default:
		return nil, ErrUnsupportedSeedingMethod
	}

	now := time.Now().UTC()
	result := make([]*Seeding, len(regs))

	err = s.db.InTransaction(ctx, func(txCtx context.Context) error {
		return s.db.WithDbSession(txCtx, func(sess db.DBSession) error {
			// Re-seeding replaces any existing seeds for this division.
			_, err := sess.Exec(txCtx, `DELETE FROM seedings WHERE division_id = $1`, cmd.DivisionID)
			if err != nil {
				return err
			}

			for seedNum, idx := range order {
				id, err := uuid.NewV7()
				if err != nil {
					return err
				}

				sd := &Seeding{
					ID:             id,
					DivisionID:     cmd.DivisionID,
					RegistrationID: regs[idx].ID,
					SeedNumber:     seedNum + 1,
					Method:         cmd.Method,
					SeededBy:       cmd.SeededBy,
					SeededAt:       now,
					CreatedAt:      now,
				}
				result[seedNum] = sd

				_, err = sess.Exec(txCtx, `
					INSERT INTO seedings (id, division_id, registration_id, seed_number, method, seeded_by, seeded_at, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`, sd.ID, sd.DivisionID, sd.RegistrationID, sd.SeedNumber, sd.Method, sd.SeededBy, sd.SeededAt, sd.CreatedAt)
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

	return result, nil
}
