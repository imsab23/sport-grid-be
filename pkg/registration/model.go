package registration

import (
	"time"

	"github.com/google/uuid"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	ErrRegistrationNotFound    = apperror.New("REG0000", "Registration not found.")
	ErrRegistrationNotOpen     = apperror.New("REG0001", "Tournament is not open for registration.")
	ErrDivisionNotActive       = apperror.New("REG0002", "Division is not active.")
	ErrDivisionFull            = apperror.New("REG0003", "Division has reached its registration capacity.")
	ErrAlreadyRegistered       = apperror.New("REG0004", "Player is already registered for this division.")
	ErrInvalidRegistrationOp   = apperror.New("REG0005", "Registration cannot be modified in its current state.")
	ErrPaymentNotFound         = apperror.New("PAY0000", "Payment submission not found.")
	ErrPaymentMethodRequired   = apperror.New("PAY0001", "Payment method is required.")
	ErrInvalidAmountSubmitted  = apperror.New("PAY0002", "Amount submitted must be greater than zero.")
	ErrInvalidPaymentOp        = apperror.New("PAY0003", "Payment submission cannot be modified in its current state.")
	ErrRejectionReasonRequired = apperror.New("PAY0004", "Rejection reason is required.")
)

const (
	RegistrationTable      = "tournament_registrations"
	PaymentSubmissionTable = "payment_submissions"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusCancelled Status = "CANCELLED"
)

type PaymentStatus string

const (
	PaymentStatusNotSubmitted PaymentStatus = "NOT_SUBMITTED"
	PaymentStatusSubmitted    PaymentStatus = "SUBMITTED"
	PaymentStatusVerified     PaymentStatus = "VERIFIED"
	PaymentStatusRejected     PaymentStatus = "REJECTED"
)

type Registration struct {
	ID            uuid.UUID     `db:"id"             json:"id"`
	TournamentID  uuid.UUID     `db:"tournament_id"  json:"tournament_id"`
	DivisionID    uuid.UUID     `db:"division_id"    json:"division_id"`
	PlayerID      uuid.UUID     `db:"player_id"      json:"player_id"`
	Status        Status        `db:"status"         json:"status"`
	PaymentStatus PaymentStatus `db:"payment_status" json:"payment_status"`
	Notes         *string       `db:"notes"          json:"notes,omitempty"`
	RegisteredAt  time.Time     `db:"registered_at"  json:"registered_at"`
	ApprovedAt    *time.Time    `db:"approved_at"    json:"approved_at,omitempty"`
	ApprovedBy    *uuid.UUID    `db:"approved_by"    json:"approved_by,omitempty"`
	CreatedAt     time.Time     `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"     json:"updated_at"`
}

// CreateRegistrationCommand.TournamentID/DivisionID are set by the handler from the URL path;
// PlayerID is set by the handler from the authenticated player identity. Neither is bindable from the body.
type CreateRegistrationCommand struct {
	TournamentID uuid.UUID `json:"-"`
	DivisionID   uuid.UUID `json:"-"`
	PlayerID     uuid.UUID `json:"-"`
	Notes        *string   `json:"notes"`
}

type SearchRegistrationQuery struct {
	TournamentID *uuid.UUID `db:"tournament_id" query:"-"`
	DivisionID   *uuid.UUID `db:"division_id"   query:"division_id"   empty:"skip"`
	PlayerID     *uuid.UUID `db:"player_id"      query:"-"`
	Status       *Status    `db:"status"         query:"status"       empty:"skip"`
	Meta         *meta.Meta
}

type SearchRegistrationResult struct {
	Registrations []*Registration `json:"registrations"`
	Meta          *meta.Meta      `json:"meta"`
}

type AmountStatus string

const (
	AmountExact     AmountStatus = "EXACT"
	AmountUnderpaid AmountStatus = "UNDERPAID"
	AmountOverpaid  AmountStatus = "OVERPAID"
)

type PaymentSubmissionStatus string

const (
	PaymentSubmissionSubmitted PaymentSubmissionStatus = "SUBMITTED"
	PaymentSubmissionVerified  PaymentSubmissionStatus = "VERIFIED"
	PaymentSubmissionRejected  PaymentSubmissionStatus = "REJECTED"
)

type PaymentSubmission struct {
	ID              uuid.UUID               `db:"id"                json:"id"`
	RegistrationID  uuid.UUID               `db:"registration_id"   json:"registration_id"`
	AmountExpected  float64                 `db:"amount_expected"   json:"amount_expected"`
	AmountSubmitted float64                 `db:"amount_submitted"  json:"amount_submitted"`
	PaymentMethod   string                  `db:"payment_method"    json:"payment_method"`
	ReferenceNumber *string                 `db:"reference_number"  json:"reference_number,omitempty"`
	ProofReference  *string                 `db:"proof_reference"   json:"proof_reference,omitempty"`
	Status          PaymentSubmissionStatus `db:"status"            json:"status"`
	AmountStatus    AmountStatus            `db:"amount_status"     json:"amount_status"`
	SubmittedAt     time.Time               `db:"submitted_at"      json:"submitted_at"`
	VerifiedAt      *time.Time              `db:"verified_at"       json:"verified_at,omitempty"`
	VerifiedBy      *uuid.UUID              `db:"verified_by"       json:"verified_by,omitempty"`
	RejectionReason *string                 `db:"rejection_reason"  json:"rejection_reason,omitempty"`
	CreatedAt       time.Time               `db:"created_at"        json:"created_at"`
	UpdatedAt       time.Time               `db:"updated_at"        json:"updated_at"`
}

// CreatePaymentSubmissionCommand.RegistrationID is set by the handler from the URL path.
type CreatePaymentSubmissionCommand struct {
	RegistrationID  uuid.UUID `json:"-"`
	AmountSubmitted float64   `json:"amount_submitted"`
	PaymentMethod   string    `json:"payment_method"`
	ReferenceNumber *string   `json:"reference_number"`
	ProofReference  *string   `json:"proof_reference"`
}

func (cmd *CreatePaymentSubmissionCommand) Validate() error {
	if !validate.RequiredString(cmd.PaymentMethod) {
		return ErrPaymentMethodRequired
	}
	if cmd.AmountSubmitted <= 0 {
		return ErrInvalidAmountSubmitted
	}
	return nil
}

type RejectPaymentSubmissionCommand struct {
	ID     uuid.UUID
	Reason string `json:"reason"`
}

func (cmd *RejectPaymentSubmissionCommand) Validate() error {
	if !validate.RequiredString(cmd.Reason) {
		return ErrRejectionReasonRequired
	}
	return nil
}
