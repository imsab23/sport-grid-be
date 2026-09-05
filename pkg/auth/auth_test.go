package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/user"

	"github.com/google/uuid"

	"github.com/imsab23/platform-be/pkg/security/jwt"
	"github.com/imsab23/platform-be/pkg/session"
	"github.com/imsab23/platform-be/pkg/session/memory"
)

// ── fakes ────────────────────────────────────────────────────────────────

type fakeUserService struct {
	byEmail map[string]*user.User
	byID    map[string]*user.User
}

func (f *fakeUserService) Search(context.Context, *user.SearchUserQuery) (*user.SearchUserResult, error) {
	return nil, nil
}

func (f *fakeUserService) Create(context.Context, *user.CreateUserCommand) (*user.User, error) {
	return nil, nil
}

func (f *fakeUserService) GetByID(_ context.Context, id string) (*user.User, error) {
	return f.byID[id], nil
}

func (f *fakeUserService) GetByEmail(_ context.Context, email string) (*user.User, error) {
	return f.byEmail[email], nil
}

func (f *fakeUserService) UpdateLastLogin(context.Context, uuid.UUID) error { return nil }

// ValidatePassword treats PasswordHash as the literal expected password — real
// hashing is the platform password package's responsibility, not auth's.
func (f *fakeUserService) ValidatePassword(_ context.Context, rawPassword, hash string) error {
	if rawPassword != hash {
		return errors.New("invalid password")
	}
	return nil
}

type fakePlayerService struct{}

func (f *fakePlayerService) Search(context.Context, *player.SearchPlayerQuery) (*player.SearchPlayerResult, error) {
	return nil, nil
}

func (f *fakePlayerService) Create(context.Context, *player.CreatePlayerCommand) (*player.Player, error) {
	return nil, nil
}

func (f *fakePlayerService) GetByID(context.Context, uuid.UUID) (*player.Player, error) {
	return nil, nil
}

func (f *fakePlayerService) GetByEmail(context.Context, string) (*player.Player, error) {
	return nil, nil
}

func (f *fakePlayerService) Update(context.Context, *player.UpdatePlayerCommand) (*player.Player, error) {
	return nil, nil
}

func (f *fakePlayerService) ValidatePassword(context.Context, string, string) error { return nil }

func (f *fakePlayerService) UpdateLastLogin(context.Context, uuid.UUID) error { return nil }

// ── test setup ─────────────────────────────────────────────────────────────

const testPassword = "correct-password"

// newTestService builds a *service directly, bypassing Init/FileKeyLoader disk
// I/O — key loading itself is platform-be's responsibility and already covered
// by its own tests.
func newTestService(t *testing.T) (*service, *fakeUserService, *user.User) {
	t.Helper()

	cfg := jwt.DefaultConfig("test-issuer", "test-audience")
	cfg.AccessTokenTTL = 15 * time.Minute
	cfg.RefreshTokenTTL = time.Hour

	kp, err := jwt.GenerateEdDSAKeyPair("k1")
	if err != nil {
		t.Fatalf("GenerateEdDSAKeyPair: %v", err)
	}
	provider, err := jwt.NewStaticKeyProvider(kp)
	if err != nil {
		t.Fatalf("NewStaticKeyProvider: %v", err)
	}
	signer, err := jwt.NewSigner(cfg, provider)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	verifier, err := jwt.NewVerifier(cfg, provider)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}

	store := memory.NewStore(session.ManagerConfig{DefaultTTL: cfg.RefreshTokenTTL, CleanupInterval: time.Minute})
	t.Cleanup(store.Close)
	sessions, err := session.NewManager(store, session.ManagerConfig{DefaultTTL: cfg.RefreshTokenTTL})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	testUser := &user.User{
		ID:           uuid.New(),
		Email:        "a@b.com",
		PasswordHash: testPassword,
		FirstName:    "Test",
	}

	fakeUsers := &fakeUserService{
		byEmail: map[string]*user.User{testUser.Email: testUser},
		byID:    map[string]*user.User{testUser.ID.String(): testUser},
	}

	svc := &service{
		signer:    signer,
		verifier:  verifier,
		jwtCfg:    cfg,
		userSvc:   fakeUsers,
		playerSvc: &fakePlayerService{},
		sessions:  sessions,
	}

	return svc, fakeUsers, testUser
}

// ── login ────────────────────────────────────────────────────────────────

func TestLoginUser_Success(t *testing.T) {
	svc, _, testUser := newTestService(t)
	ctx := context.Background()

	result, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: testPassword, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.ExpiresAt == 0 {
		t.Error("expected ExpiresAt to be populated")
	}

	stored, err := svc.sessions.GetByIndex(ctx, session.IndexRefreshToken, jwt.HashRefreshToken(result.RefreshToken))
	if err != nil {
		t.Fatalf("GetByIndex: %v", err)
	}
	if stored.Metadata[session.MetadataKeyClientIP] != "203.0.113.1" {
		t.Errorf("expected stored client IP 203.0.113.1, got %q", stored.Metadata[session.MetadataKeyClientIP])
	}
	if strings.Contains(string(stored.Data), result.RefreshToken) {
		t.Error("plaintext refresh token must not be stored in session data")
	}
}

func TestLoginUser_InvalidPasswordRejected(t *testing.T) {
	svc, _, testUser := newTestService(t)
	ctx := context.Background()

	_, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: "wrong", ClientIP: "203.0.113.1"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ── refresh ──────────────────────────────────────────────────────────────

func TestRefreshToken_SuccessRotatesAndInvalidatesOldToken(t *testing.T) {
	svc, _, testUser := newTestService(t)
	ctx := context.Background()

	login, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: testPassword, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	refreshed, err := svc.RefreshToken(ctx, &RefreshTokenCommand{RefreshToken: login.RefreshToken, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Error("expected new access and refresh tokens")
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Error("expected a rotated (different) refresh token")
	}
	if refreshed.ExpiresAt == 0 {
		t.Error("expected ExpiresAt to be populated")
	}

	if _, err := svc.RefreshToken(ctx, &RefreshTokenCommand{RefreshToken: login.RefreshToken, ClientIP: "203.0.113.1"}); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken reusing rotated token, got %v", err)
	}
}

func TestRefreshToken_IPMismatchRejected(t *testing.T) {
	svc, _, testUser := newTestService(t)
	ctx := context.Background()

	login, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: testPassword, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	_, err = svc.RefreshToken(ctx, &RefreshTokenCommand{RefreshToken: login.RefreshToken, ClientIP: "198.51.100.9"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken on IP mismatch, got %v", err)
	}
}

func TestRefreshToken_MissingTokenRejected(t *testing.T) {
	svc, _, _ := newTestService(t)
	_, err := svc.RefreshToken(context.Background(), &RefreshTokenCommand{ClientIP: "203.0.113.1"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for missing token, got %v", err)
	}
}

func TestRefreshToken_UnknownTokenRejected(t *testing.T) {
	svc, _, _ := newTestService(t)
	_, err := svc.RefreshToken(context.Background(), &RefreshTokenCommand{RefreshToken: "never-issued", ClientIP: "203.0.113.1"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for unknown token, got %v", err)
	}
}

func TestRefreshToken_DeletedUserRejected(t *testing.T) {
	svc, fakeUsers, testUser := newTestService(t)
	ctx := context.Background()

	login, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: testPassword, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	delete(fakeUsers.byID, testUser.ID.String())

	_, err = svc.RefreshToken(ctx, &RefreshTokenCommand{RefreshToken: login.RefreshToken, ClientIP: "203.0.113.1"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for deleted user, got %v", err)
	}
}

func TestRefreshToken_ConcurrentReuseAllowsOnlyOneSuccess(t *testing.T) {
	svc, _, testUser := newTestService(t)
	ctx := context.Background()

	login, err := svc.LoginUser(ctx, &LoginUserCommand{Email: testUser.Email, Password: testPassword, ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}

	const attempts = 5
	var wg sync.WaitGroup
	var successes int32
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.RefreshToken(ctx, &RefreshTokenCommand{RefreshToken: login.RefreshToken, ClientIP: "203.0.113.1"}); err == nil {
				atomic.AddInt32(&successes, 1)
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Errorf("expected exactly 1 successful rotation, got %d", successes)
	}
}
