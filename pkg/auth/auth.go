package auth

import (
	"context"
	"sport-grid-be/pkg/config"
	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/user"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	"github.com/imsab23/platform-be/pkg/security/jwt"
)

type Service interface {
	Register(ctx context.Context, cmd *RegisterUserCommand) error
	LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error)
	VerifyToken() *jwt.Verifier
}

type service struct {
	signer    *jwt.Signer
	verifier  *jwt.Verifier
	jwtCfg    jwt.Config
	userSvc   user.Service
	playerSvc player.Service
	db        db.DB
}

func NewService(userSvc user.Service, playerSvc player.Service, database db.DB, jwtCfg *config.JWTConfig) (Service, error) {
	cfg, signer, verifier, err := Init(jwtCfg)
	if err != nil {
		return nil, err
	}

	return &service{
		signer:    signer,
		verifier:  verifier,
		jwtCfg:    cfg,
		userSvc:   userSvc,
		playerSvc: playerSvc,
		db:        database,
	}, nil
}

// Register creates a user and player profile atomically within one transaction.
func (s *service) Register(ctx context.Context, cmd *RegisterUserCommand) error {
	return s.db.InTransaction(ctx, func(ctx context.Context) error {
		u, err := s.userSvc.Create(ctx, &user.CreateUserCommand{
			Email:    cmd.Email,
			Password: cmd.Password,
		})
		if err != nil {
			return err
		}

		_, err = s.playerSvc.Create(ctx, &player.CreatePlayerCommand{
			UserID:     u.ID,
			FirstName:  cmd.FirstName,
			LastName:   cmd.LastName,
			MiddleName: cmd.MiddleName,
			Phone:      cmd.Phone,
		})
		return err
	})
}

func (s *service) LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error) {
	existing, err := s.userSvc.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, ErrInvalidCredentials
	}

	err = s.userSvc.ValidatePassword(ctx, cmd.Password, existing.PasswordHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Best-effort; login must not fail if the timestamp update fails.
	_ = s.userSvc.UpdateLastLogin(ctx, existing.ID)

	accessToken, err := s.GenerateToken(&UserAuth{
		ID:       existing.ID.String(),
		Role:     existing.Role,
		UserType: User,
		ClientID: existing.ClientID.String(),
	})
	if err != nil {
		return nil, err
	}

	return &LoginUserResult{
		FirstName:   existing.FirstName,
		AccessToken: accessToken,
	}, nil
}
