package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	"github.com/imsab23/platform-be/infra/storage/postgres/helper"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	"github.com/imsab23/platform-be/pkg/security/password"
)

type Service interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ValidatePassword(ctx context.Context, password, encryptedPassword string) error
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

func (s *service) Create(ctx context.Context, user *User) error {
	isExists, err := s.GetByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	if isExists != nil {
		return ErrUserAlreadyExists
	}

	user.Password, err = s.hasher.Hash(user.Password)
	if err != nil {
		return err
	}

	err = s.db.InTransaction(ctx, func(ctx context.Context) error {
		_, err := helperdb.Create(ctx, s.db, UserTable, user)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	user := &User{}

	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	err = helper.GetByField(ctx, s.db, UserTable, &User{ID: userID}, user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	var result User
	err := helper.GetByField(ctx, s.db, UserTable, &User{Email: email}, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &result, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateUserCommand) error {
	_, err := s.GetByID(ctx, cmd.ID.String())
	if err != nil {
		return err
	}

	return s.db.InTransaction(ctx, func(ctx context.Context) error {
		err := helper.Update(ctx, s.db, UserTable, cmd.ID, &User{
			FirstName:     cmd.FirstName,
			MiddleName:    cmd.MiddleName,
			LastName:      cmd.LastName,
			Email:         cmd.Email,
			ContactNumber: cmd.ContactNumber,
			UpdatedBy:     cmd.UpdatedBy,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *service) Search(ctx context.Context, query *SearchUserQuery) ([]*SearchUserResult, error) {

	return nil, nil
}

func (s *service) ValidatePassword(ctx context.Context, password, encryptedPassword string) error {
	return s.hasher.Verify(password, encryptedPassword)
}
