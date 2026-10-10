package registration

import (
	"context"

	"github.com/google/uuid"
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
