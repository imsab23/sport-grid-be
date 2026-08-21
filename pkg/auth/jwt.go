package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/imsab23/platform-be/pkg/security/jwt"
)

func loadEdDSAKeyPair(
	kid string,
	privateKeyPath string,
	publicKeyPath string,
) (jwt.KeyPair, error) {
	// Load private key.
	privatePEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: read private key: %w", err)
	}

	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: parse private key: %w", err)
	}

	edPrivateKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return jwt.KeyPair{}, fmt.Errorf("jwt: private key is not Ed25519")
	}

	// Load public key.
	publicPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: read public key: %w", err)
	}

	publicBlock, _ := pem.Decode(publicPEM)
	if publicBlock == nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: decode public key PEM")
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return jwt.KeyPair{}, fmt.Errorf("jwt: parse public key: %w", err)
	}

	edPublicKey, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return jwt.KeyPair{}, fmt.Errorf("jwt: public key is not Ed25519")
	}

	return jwt.KeyPair{
		ID:         kid,
		PrivateKey: edPrivateKey,
		PublicKey:  edPublicKey,
	}, nil
}

func Init() (cfg jwt.Config, signer *jwt.Signer, verifier *jwt.Verifier, err error) {
	cfg = jwt.DefaultConfig(
		"sport-grid",
		"sport-grid",
	)

	kp, err := loadEdDSAKeyPair(
		"sport-grid",
		"/home/naeem/Documents/Sam/sport-grid-be/keys/jwt_private.pem",
		"/home/naeem/Documents/Sam/sport-grid-be/keys/jwt_public.pem",
	)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("load JWT key pair: %w", err)
	}

	provider, err := jwt.NewStaticKeyProvider(kp)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("create key provider: %w", err)
	}

	signer, err = jwt.NewSigner(cfg, provider)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("create signer: %w", err)
	}

	verifier, err = jwt.NewVerifier(cfg, provider)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("create verifier: %w", err)
	}

	return cfg, signer, verifier, nil
}

func (s *service) GenerateToken(user *UserAuth) (string, error) {
	now := time.Now().UTC()
	claims := jwt.NewAccessClaims(s.jwtCfg, user.ID, user.ID, "user", user.ClientID, []string{user.Role}, nil, now)

	accessToken, err := s.signer.Sign(claims)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return accessToken, nil
}

func (s *service) VerifyToken() *jwt.Verifier {
	return s.verifier
}
