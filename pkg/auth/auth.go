package auth

import (
	"context"
	"sport-grid-be/pkg/user"

	"github.com/imsab23/platform-be/pkg/security/jwt"
)

type Service interface {
	LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error)
	VerifyToken() *jwt.Verifier
}

type service struct {
	signer   *jwt.Signer
	verifier *jwt.Verifier
	jwtCfg   jwt.Config
	userSvc  user.Service
}

func NewService(userSvc user.Service) (Service, error) {
	jwtCfg, signer, verifier, err := Init()
	if err != nil {
		return nil, err
	}

	return &service{
		signer:   signer,
		verifier: verifier,
		jwtCfg:   jwtCfg,
		userSvc:  userSvc,
	}, nil
}

func (s *service) LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error) {
	isExists, err := s.userSvc.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if isExists == nil {
		return nil, ErrInvalidCredentials
	}

	err = s.userSvc.ValidatePassword(ctx, cmd.Password, isExists.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.GenerateToken(&UserAuth{
		ID: isExists.ID.String(),
		// Role:     isExists.Role,
		// UserType: isExists.UserType,
		// ClientID: isExists.ClientID,
	})
	if err != nil {
		return nil, err
	}

	return &LoginUserResult{
		Token: accessToken,
	}, nil
}
