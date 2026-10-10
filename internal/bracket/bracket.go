package bracket

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Generate(ctx context.Context, cmd *GenerateBracketCommand) (*Bracket, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Bracket, error)
	GetNodes(ctx context.Context, bracketID uuid.UUID) ([]*BracketNode, error)
	Reset(ctx context.Context, cmd *ResetBracketCommand) (*Bracket, error)
	AdvanceWinner(ctx context.Context, nodeID uuid.UUID, registrationID uuid.UUID) error
}
