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
)

type Service interface {
	Search(ctx context.Context, query *SearchPlayerQuery) (*SearchPlayerResult, error)
	Create(ctx context.Context, cmd *CreatePlayerCommand) (*Player, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*Player, error)
	GetByPlayerID(ctx context.Context, id uuid.UUID) (*Player, error)
	Update(ctx context.Context, cmd *UpdatePlayerCommand) (*Player, error)
}

type service struct {
	db db.DB
}

func NewService(db db.DB) (Service, error) {
	return &service{db: db}, nil
}

func (s *service) Create(ctx context.Context, cmd *CreatePlayerCommand) (*Player, error) {
	now := time.Now().UTC()
	profile := &Player{
		UserID:     cmd.UserID,
		FirstName:  cmd.FirstName,
		MiddleName: cmd.MiddleName,
		LastName:   cmd.LastName,
		Phone:      cmd.Phone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := helperdb.Create(ctx, s.db, PlayerTable, profile, helperdb.CreateOptions{
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

func (s *service) GetByID(ctx context.Context, userID uuid.UUID) (*Player, error) {
	var profile Player
	err := helperdb.GetByField(ctx, s.db, PlayerTable, &Player{UserID: userID}, &profile)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlayerProfileNotFound
		}
		return nil, err
	}

	return &profile, nil
}

func (s *service) GetByPlayerID(ctx context.Context, id uuid.UUID) (*Player, error) {
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

func (s *service) Update(ctx context.Context, cmd *UpdatePlayerCommand) (*Player, error) {
	profile, err := s.GetByID(ctx, cmd.UserID)
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

// playerJoinTable is the FROM expression that resolves player fields alongside user email.
const playerJoinTable = "players INNER JOIN users ON players.user_id = users.id"

func (s *service) Search(ctx context.Context, query *SearchPlayerQuery) (*SearchPlayerResult, error) {
	sortDir := searchHelper.SortAsc
	if strings.EqualFold(query.Meta.Order, "desc") {
		sortDir = searchHelper.SortDesc
	}

	params := searchHelper.Params{
		Columns: []string{
			"players.id AS id",
			"players.user_id AS user_id",
			"players.first_name AS first_name",
			"players.middle_name AS middle_name",
			"players.last_name AS last_name",
			"players.phone AS phone",
			"players.created_at AS created_at",
			"users.email AS email",
		},
		Search:      query.Search,
		Searchable:  []string{"players.first_name", "players.last_name", "users.email"},
		SortBy:      query.Meta.OrderBy,
		SortDir:     sortDir,
		CursorField: "players.id",
	}

	result, err := searchHelper.Search[*PlayerSearchItem](ctx, s.db, searchHelper.From{Table: playerJoinTable}, params)
	if err != nil {
		return nil, err
	}

	return &SearchPlayerResult{
		Players: result.Items,
		Meta:    result.ToMeta(query.Meta.OrderBy, sortDir),
	}, nil
}
