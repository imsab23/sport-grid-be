package player

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	helperdb "github.com/imsab23/platform-be/infra/storage/postgres/helper"
	searchHelper "github.com/imsab23/platform-be/infra/storage/postgres/helper/query"
	"github.com/imsab23/platform-be/pkg/security/password"
)

type Service interface {
	Search(ctx context.Context, query *SearchPlayerQuery) (*SearchPlayerResult, error)
	Create(ctx context.Context, cmd *CreatePlayerCommand) (*Player, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Player, error)
	GetByEmail(ctx context.Context, email string) (*Player, error)
	Update(ctx context.Context, cmd *UpdatePlayerCommand) (*Player, error)
	ValidatePassword(ctx context.Context, rawPassword, hash string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
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

	return &service{db: db, hasher: hasher}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreatePlayerCommand) (*Player, error) {
	existing, err := s.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPlayerAlreadyExists
	}

	hash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	profile := &Player{
		Email:        cmd.Email,
		PasswordHash: hash,
		FirstName:    cmd.FirstName,
		MiddleName:   cmd.MiddleName,
		LastName:     cmd.LastName,
		Phone:        cmd.Phone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = helperdb.Create(ctx, s.db, PlayerTable, profile, helperdb.CreateOptions{
		ID: helperdb.IDOptions{
			Mode:  helperdb.IDApplication,
			Force: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Player, error) {
	var profile Player
	err := helperdb.GetByField(ctx, s.db, PlayerTable, &Player{ID: id}, &profile)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlayerProfileNotFound
		}
		return nil, err
	}

	return &profile, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*Player, error) {
	var profile Player
	err := helperdb.GetByField(ctx, s.db, PlayerTable, &Player{Email: email}, &profile)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *service) ValidatePassword(_ context.Context, rawPassword, hash string) error {
	return s.hasher.Verify(rawPassword, hash)
}

// playerLastLoginUpdate is a targeted struct so UpdateLastLogin only touches the two columns.
type playerLastLoginUpdate struct {
	LastLoginAt time.Time `db:"last_login_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (s *service) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return helperdb.Update(ctx, s.db, PlayerTable, id, &playerLastLoginUpdate{
		LastLoginAt: now,
		UpdatedAt:   now,
	})
}

func (s *service) Update(ctx context.Context, cmd *UpdatePlayerCommand) (*Player, error) {
	profile, err := s.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	updateData := &PlayerUpdate{
		FirstName:             cmd.FirstName,
		MiddleName:            cmd.MiddleName,
		LastName:              cmd.LastName,
		DateOfBirth:           cmd.DateOfBirth,
		Gender:                cmd.Gender,
		Phone:                 cmd.Phone,
		Address:               cmd.Address,
		EmergencyContactName:  cmd.EmergencyContactName,
		EmergencyContactPhone: cmd.EmergencyContactPhone,
		UpdatedAt:             now,
	}

	err = helperdb.Update(ctx, s.db, PlayerTable, profile.ID, updateData)
	if err != nil {
		return nil, err
	}

	profile.FirstName = cmd.FirstName
	profile.MiddleName = cmd.MiddleName
	profile.LastName = cmd.LastName
	profile.DateOfBirth = cmd.DateOfBirth
	profile.Gender = cmd.Gender
	profile.Phone = cmd.Phone
	profile.Address = cmd.Address
	profile.EmergencyContactName = cmd.EmergencyContactName
	profile.EmergencyContactPhone = cmd.EmergencyContactPhone
	profile.UpdatedAt = now

	return profile, nil
}

func (s *service) Search(ctx context.Context, query *SearchPlayerQuery) (*SearchPlayerResult, error) {
	sortDir := searchHelper.SortAsc
	if strings.EqualFold(query.Meta.Order, "desc") {
		sortDir = searchHelper.SortDesc
	}

	params := searchHelper.Params{
		Columns: []string{
			"id",
			"first_name",
			"middle_name",
			"last_name",
			"phone",
			"email",
			"created_at",
		},
		Search:      query.Search,
		Searchable:  []string{"first_name", "last_name", "email"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     sortDir,
		CursorField: "id",
	}

	result, err := searchHelper.Search[*PlayerSearchItem](ctx, s.db, searchHelper.From{Table: PlayerTable}, params)
	if err != nil {
		return nil, err
	}

	return &SearchPlayerResult{
		Players: result.Items,
		Meta:    result.ToMeta(query.Meta.OrderBy, sortDir),
	}, nil
}
