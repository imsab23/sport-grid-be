package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"sport-grid-be/pkg/config"
	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/user"

	"github.com/google/uuid"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	"github.com/imsab23/platform-be/pkg/security/jwt"
	"github.com/imsab23/platform-be/pkg/session"
	"github.com/imsab23/platform-be/pkg/session/memory"
)

type Service interface {
	Register(ctx context.Context, cmd *RegisterUserCommand) error
	LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error)
	RefreshToken(ctx context.Context, cmd *RefreshTokenCommand) (*LoginUserResult, error)
	VerifyToken() *jwt.Verifier
}

type service struct {
	signer    *jwt.Signer
	verifier  *jwt.Verifier
	jwtCfg    jwt.Config
	userSvc   user.Service
	playerSvc player.Service
	db        db.DB
	sessions  *session.Manager
	refreshMu sync.Mutex
}

func NewService(userSvc user.Service, playerSvc player.Service, database db.DB, jwtCfg *config.JWTConfig) (Service, error) {
	sessionStore := memory.NewStore(session.ManagerConfig{
		DefaultTTL:      jwt.DefaultRefreshTokenTTL,
		CleanupInterval: time.Minute,
	})

	sessions, err := session.NewManager(sessionStore, session.ManagerConfig{
		DefaultTTL:      jwt.DefaultRefreshTokenTTL,
		CleanupInterval: time.Minute,
	})
	if err != nil {
		sessionStore.Close()
		return nil, err
	}

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
		sessions:  sessions,
	}, nil
}

// Register creates a standalone player account — players have no relation to users.
func (s *service) Register(ctx context.Context, cmd *RegisterUserCommand) error {
	_, err := s.playerSvc.Create(ctx, &player.CreatePlayerCommand{
		Email:      cmd.Email,
		Password:   cmd.Password,
		FirstName:  cmd.FirstName,
		LastName:   cmd.LastName,
		MiddleName: cmd.MiddleName,
		Phone:      cmd.Phone,
	})
	return err
}

func (s *service) LoginUser(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error) {
	existingUser, err := s.userSvc.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		err = s.userSvc.ValidatePassword(ctx, cmd.Password, existingUser.PasswordHash)
		if err != nil {
			return nil, ErrInvalidCredentials
		}

		// Best-effort; login must not fail if the timestamp update fails.
		_ = s.userSvc.UpdateLastLogin(ctx, existingUser.ID)

		clientID := ""
		if existingUser.ClientID != nil {
			clientID = existingUser.ClientID.String()
		}

		result, err := s.issueTokens(ctx, &UserAuth{
			ID:       existingUser.ID.String(),
			Role:     existingUser.Role,
			UserType: User,
			ClientID: clientID,
		}, cmd.ClientIP, existingUser.FirstName)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	existingPlayer, err := s.playerSvc.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if existingPlayer == nil {
		return nil, ErrInvalidCredentials
	}

	err = s.playerSvc.ValidatePassword(ctx, cmd.Password, existingPlayer.PasswordHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Best-effort; login must not fail if the timestamp update fails.
	_ = s.playerSvc.UpdateLastLogin(ctx, existingPlayer.ID)

	result, err := s.issueTokens(ctx, &UserAuth{
		ID:       existingPlayer.ID.String(),
		UserType: Player,
	}, cmd.ClientIP, existingPlayer.FirstName)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type refreshSessionData struct {
	UserID   string   `json:"user_id"`
	UserType UserType `json:"user_type"`
}

func (s *service) issueTokens(ctx context.Context, userAuth *UserAuth, clientIP, firstName string) (*LoginUserResult, error) {
	accessToken, claims, err := s.generateAccessToken(userAuth, clientIP)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	data, err := json.Marshal(refreshSessionData{UserID: userAuth.ID, UserType: userAuth.UserType})
	if err != nil {
		return nil, fmt.Errorf("encode refresh session: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.sessions.Create(ctx, userAuth.ID, "refresh", data,
		map[string]string{session.MetadataKeyClientIP: clientIP},
		map[string]string{session.IndexRefreshToken: jwt.HashRefreshToken(refreshToken)})
	if err != nil {
		return nil, fmt.Errorf("store refresh session: %w", err)
	}

	return &LoginUserResult{
		FirstName:        firstName,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        claims.ExpiresAt.Time.Unix(),
		RefreshExpiresAt: now.Add(s.jwtCfg.RefreshTokenTTL).Unix(),
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, cmd *RefreshTokenCommand) (*LoginUserResult, error) {
	if cmd == nil || cmd.RefreshToken == "" || cmd.ClientIP == "" {
		return nil, ErrInvalidRefreshToken
	}

	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	stored, err := s.sessions.GetByIndex(ctx, session.IndexRefreshToken, jwt.HashRefreshToken(cmd.RefreshToken))
	if err != nil || stored.Metadata[session.MetadataKeyClientIP] != cmd.ClientIP {
		return nil, ErrInvalidRefreshToken
	}

	if stored.IsExpired() {
		_ = s.sessions.Revoke(ctx, stored.ID)
		return nil, ErrInvalidRefreshToken
	}

	var data refreshSessionData
	if err := json.Unmarshal(stored.Data, &data); err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userAuth, firstName, err := s.loadUserAuth(ctx, data)
	if err != nil {
		_ = s.sessions.Revoke(ctx, stored.ID)
		return nil, ErrInvalidRefreshToken
	}

	// Revoke before issuing the replacement so a second concurrent use fails.
	err = s.sessions.Revoke(ctx, stored.ID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	return s.issueTokens(ctx, userAuth, cmd.ClientIP, firstName)
}

func (s *service) loadUserAuth(ctx context.Context, data refreshSessionData) (*UserAuth, string, error) {
	switch data.UserType {
	case User:
		u, err := s.userSvc.GetByID(ctx, data.UserID)
		if err != nil || u == nil {
			return nil, "", ErrInvalidRefreshToken
		}
		clientID := ""
		if u.ClientID != nil {
			clientID = u.ClientID.String()
		}
		return &UserAuth{ID: u.ID.String(), Role: u.Role, UserType: User, ClientID: clientID}, u.FirstName, nil
	case Player:
		id, err := uuid.Parse(data.UserID)
		if err != nil {
			return nil, "", ErrInvalidRefreshToken
		}
		p, err := s.playerSvc.GetByID(ctx, id)
		if err != nil || p == nil {
			return nil, "", ErrInvalidRefreshToken
		}
		return &UserAuth{ID: p.ID.String(), UserType: Player}, p.FirstName, nil
	default:
		return nil, "", ErrInvalidRefreshToken
	}
}
