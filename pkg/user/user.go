package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
	"github.com/imsab23/platform-be/pkg/security/password"
	"github.com/imsab23/platform-be/pkg/util"
)

type Service interface {
	Search(ctx context.Context, query *SearchUserQuery) (*SearchUserResult, error)
	Create(ctx context.Context, cmd *CreateUserCommand) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	// Update(ctx context.Context, cmd *UpdateUserCommand) (*User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	ValidatePassword(ctx context.Context, rawPassword, hash string) error
}

type service struct {
	db     db.DB
	hasher *password.Hasher
}

func NewService(db db.DB) (Service, error) {
	hasher, err := password.NewHasher(password.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &service{
		db:     db,
		hasher: hasher,
	}, nil
}

func (s *service) Search(ctx context.Context, query *SearchUserQuery) (*SearchUserResult, error) {
	params := searchHelper.Params{
		Columns:     []string{"u.id"},
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"email"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchHelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	result, err := searchHelper.Search[*User](ctx, s.db, searchHelper.From{Table: UserTable}, params)
	if err != nil {
		return nil, err
	}

	m := result.ToMeta(query.Meta.OrderBy, searchHelper.SortDirection(query.Meta.Order))
	return &SearchUserResult{
		Users: result.Items,
		Meta:  &m,
	}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateUserCommand) (*User, error) {
	existing, err := s.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	entity := &User{}
	err = util.MapByJSON(cmd, entity)
	if err != nil {
		return nil, err
	}

	entity.PasswordHash = hash

	_, err = helperdb.Create(ctx, s.db, UserTable, entity, helperdb.CreateOptions{
		ID: helperdb.IDOptions{
			Mode:  helperdb.IDApplication,
			Force: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var u User
	err = helperdb.GetByField(ctx, s.db, UserTable, &User{ID: userID}, &u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := helperdb.GetByField(ctx, s.db, UserTable, &User{Email: email}, &u)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *service) ValidatePassword(_ context.Context, rawPassword, hash string) error {
	return s.hasher.Verify(rawPassword, hash)
}

// userLastLoginUpdate is a targeted struct so UpdateLastLogin only touches the two columns.
type userLastLoginUpdate struct {
	LastLoginAt time.Time `db:"last_login_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (s *service) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return helperdb.Update(ctx, s.db, UserTable, id, &userLastLoginUpdate{
		LastLoginAt: now,
		UpdatedAt:   now,
	})
}
