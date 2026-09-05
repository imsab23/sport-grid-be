package auth

import (
	"context"
	"fmt"
	"time"

	"sport-grid-be/pkg/config"

	"github.com/imsab23/platform-be/pkg/security/jwt"
)

func Init(cfg *config.JWTConfig) (jwtCfg jwt.Config, signer *jwt.Signer, verifier *jwt.Verifier, err error) {
	if cfg == nil {
		return jwt.Config{}, nil, nil, fmt.Errorf("JWT config must not be nil")
	}

	jwtCfg = jwt.DefaultConfig(cfg.Issuer, cfg.Audience)
	jwtCfg.AccessTokenTTL = cfg.AccessTokenTTL
	jwtCfg.RefreshTokenTTL = cfg.RefreshTokenTTL

	loader, err := jwt.NewFileKeyLoader(jwt.KeyConfig{
		KID:            cfg.KID,
		PrivateKeyPath: cfg.PrivateKeyPath,
		PublicKeyPath:  cfg.PublicKeyPath,
	})
	if err != nil {
		return jwtCfg, nil, nil, fmt.Errorf("create JWT key loader: %w", err)
	}

	kp, err := loader.Load(context.Background())
	if err != nil {
		return jwtCfg, nil, nil, fmt.Errorf("load JWT key pair: %w", err)
	}

	provider, err := jwt.NewStaticKeyProvider(kp)
	if err != nil {
		return jwtCfg, nil, nil, fmt.Errorf("create key provider: %w", err)
	}

	signer, err = jwt.NewSigner(jwtCfg, provider)
	if err != nil {
		return jwtCfg, nil, nil, fmt.Errorf("create signer: %w", err)
	}

	verifier, err = jwt.NewVerifier(jwtCfg, provider)
	if err != nil {
		return jwtCfg, nil, nil, fmt.Errorf("create verifier: %w", err)
	}

	return jwtCfg, signer, verifier, nil
}

func (s *service) generateAccessToken(user *UserAuth, clientIP string) (string, jwt.Claims, error) {
	now := time.Now().UTC()

	var roles []string
	if user.Role != "" {
		roles = []string{string(user.Role)}
	}

	claims := jwt.NewAccessClaims(jwt.AccessClaimsParams{
		Config:   s.jwtCfg,
		Subject:  user.ID,
		UserID:   user.ID,
		UserType: string(user.UserType),
		ClientID: user.ClientID,
		ClientIP: clientIP,
		Roles:    roles,
		Now:      now,
	})

	accessToken, err := s.signer.Sign(claims)
	if err != nil {
		return "", jwt.Claims{}, fmt.Errorf("sign access token: %w", err)
	}

	return accessToken, claims, nil
}

func (s *service) GenerateToken(user *UserAuth) (string, error) {
	accessToken, _, err := s.generateAccessToken(user, "")
	return accessToken, err
}

func (s *service) VerifyToken() *jwt.Verifier {
	return s.verifier
}
