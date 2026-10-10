package usersvc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sport-grid-be/internal/role"
	"sport-grid-be/internal/user"

	"github.com/google/uuid"
	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchhelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
	"github.com/imsab23/platform-be/pkg/security/password"
	"github.com/imsab23/platform-be/pkg/util"
)

type service struct {
	database db.DB
	hasher   *password.Hasher
}

func NewService(database db.DB) (user.Service, error) {
	hasher, err := password.NewHasher(password.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &service{
		database: database,
		hasher:   hasher,
	}, nil
}

func (service *service) Search(ctx context.Context, query *user.SearchUserQuery) (*user.SearchUserResult, error) {
	err := query.ApplyActorScope()
	if err != nil {
		return nil, err
	}

	params := searchhelper.Params{
		Filters:     query,
		Search:      &query.Search,
		Searchable:  []string{"email"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     searchhelper.SortDirection(query.Meta.Order),
		CursorField: "id",
	}

	searchResult, err := searchhelper.Search[*user.User](ctx, service.database, searchhelper.From{Table: user.UserTable}, params)
	if err != nil {
		return nil, err
	}

	searchMeta := searchResult.ToMeta(query.Meta.OrderBy, searchhelper.SortDirection(query.Meta.Order))
	return &user.SearchUserResult{
		Users: searchResult.Items,
		Meta:  &searchMeta,
	}, nil
}

func (service *service) Create(ctx context.Context, command *user.CreateUserCommand) (*user.User, error) {
	err := command.Validate()
	if err != nil {
		return nil, err
	}

	existingUser, err := service.GetByEmail(ctx, command.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, user.ErrUserAlreadyExists
	}

	passwordHash, err := service.hasher.Hash(command.Password)
	if err != nil {
		return nil, err
	}

	userRecord := &user.User{}
	err = util.MapByJSON(command, userRecord)
	if err != nil {
		return nil, err
	}

	userRecord.PasswordHash = passwordHash
	userRecord.Status = user.Active

	_, err = helperdb.Create(ctx, service.database, user.UserTable, userRecord, helperdb.CreateOptions{
		ID: helperdb.IDOptions{
			Mode:  helperdb.IDApplication,
			Force: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}

func (service *service) GetByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	var userRecord user.User
	err := helperdb.GetByField(ctx, service.database, user.UserTable, &user.User{ID: userID}, &userRecord)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &userRecord, nil
}

func (service *service) GetByIDForActor(ctx context.Context, actor role.Actor, userID uuid.UUID) (*user.User, error) {
	userRecord, err := service.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if actor.Role == role.SuperAdmin {
		return userRecord, nil
	}

	if actor.Role != role.ClientAdmin || actor.ClientID == nil || userRecord.ClientID == nil {
		return nil, user.ErrUserScopeForbidden
	}

	if *actor.ClientID != *userRecord.ClientID {
		return nil, user.ErrUserScopeForbidden
	}

	return userRecord, nil
}

func (service *service) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var userRecord user.User
	err := helperdb.GetByField(ctx, service.database, user.UserTable, &user.User{Email: email}, &userRecord)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userRecord, nil
}

func (service *service) ValidatePassword(_ context.Context, rawPassword string, passwordHash string) error {
	return service.hasher.Verify(rawPassword, passwordHash)
}

func (service *service) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return helperdb.Update(ctx, service.database, user.UserTable, userID, &user.UserLastLoginUpdate{
		LastLoginAt: now,
		UpdatedAt:   now,
	})
}

func (service *service) UpdatePassword(ctx context.Context, command *user.UpdateUserPasswordCommand) error {
	userRecord, err := service.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	err = service.hasher.Verify(command.OldPassword, userRecord.PasswordHash)
	if err != nil {
		return err
	}

	passwordHash, err := service.hasher.Hash(command.NewPassword)
	if err != nil {
		return err
	}

	command.PasswordHash = passwordHash
	return helperdb.Update(ctx, service.database, user.UserTable, command.ID, command)
}
