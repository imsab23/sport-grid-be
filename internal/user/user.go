package user

import (
	"context"

	"sport-grid-be/internal/role"

	"github.com/google/uuid"
)

type Service interface {
	Search(ctx context.Context, query *SearchUserQuery) (*SearchUserResult, error)
	Create(ctx context.Context, command *CreateUserCommand) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetByIDForActor(ctx context.Context, actor role.Actor, userID uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, command *UpdateUserPasswordCommand) error
	ValidatePassword(ctx context.Context, rawPassword string, passwordHash string) error
}
