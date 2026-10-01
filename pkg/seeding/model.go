package seeding

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrSeedingNotFound           = apperror.New("SED0000", "Seeding not found.")
	ErrInsufficientRegistrations = apperror.New("SED0001", "At least 2 confirmed registrations are required to seed a division.")
	ErrUnsupportedSeedingMethod  = apperror.New("SED0002", "Unsupported seeding method.")
	ErrSeedNumberTaken           = apperror.New("SED0003", "Seed number is already assigned to another registration.")
	ErrRegistrationNotConfirmed  = apperror.New("SED0004", "Registration must be confirmed before it can be seeded.")
	ErrInvalidSeedNumber         = apperror.New("SED0005", "Seed number must be at least 1.")
)

const (
	SeedingTable = "seedings"
)

type Method string

const (
	MethodManual            Method = "MANUAL"
	MethodRandom            Method = "RANDOM"
	MethodRegistrationOrder Method = "REGISTRATION_ORDER"
)

type Seeding struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	DivisionID     uuid.UUID `db:"division_id"     json:"division_id"`
	RegistrationID uuid.UUID `db:"registration_id" json:"registration_id"`
	SeedNumber     int       `db:"seed_number"      json:"seed_number"`
	Method         Method    `db:"method"           json:"method"`
	SeededBy       uuid.UUID `db:"seeded_by"        json:"seeded_by"`
	SeededAt       time.Time `db:"seeded_at"        json:"seeded_at"`
	CreatedAt      time.Time `db:"created_at"       json:"created_at"`
}

// AssignSeedCommand.DivisionID is set by the handler from the URL path.
type AssignSeedCommand struct {
	DivisionID     uuid.UUID `json:"-"`
	RegistrationID uuid.UUID `json:"registration_id"`
	SeedNumber     int       `json:"seed_number"`
	SeededBy       uuid.UUID `json:"-"`
}

func (cmd *AssignSeedCommand) Validate() error {
	if cmd.SeedNumber < 1 {
		return ErrInvalidSeedNumber
	}
	return nil
}

// GenerateSeedingCommand.DivisionID is set by the handler from the URL path.
type GenerateSeedingCommand struct {
	DivisionID uuid.UUID `json:"-"`
	Method     Method    `json:"method"`
	SeededBy   uuid.UUID `json:"-"`
}

func (cmd *GenerateSeedingCommand) Validate() error {
	switch cmd.Method {
	case MethodRandom, MethodRegistrationOrder:
		return nil
	default:
		return ErrUnsupportedSeedingMethod
	}
}
