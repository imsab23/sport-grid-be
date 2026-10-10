package sport

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateSportCommand) (*Sport, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Sport, error)
	GetBySlug(ctx context.Context, slug string) (*Sport, error)
	Search(ctx context.Context, query *SearchSportQuery) (*SearchSportResult, error)
	Update(ctx context.Context, cmd *UpdateSportCommand) (*Sport, error)
	UpdateStatus(ctx context.Context, cmd *UpdateStatusCommand) error
}
