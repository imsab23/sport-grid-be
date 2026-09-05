package auth

import (
	"sport-grid-be/pkg/role"

	apperror "github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrInvalidCredentials  = apperror.New("AUTH0000", "Invalid credentials.")
	ErrInvalidRefreshToken = apperror.New("AUTH0001", "Invalid or expired refresh token.")
)

type UserType string

const (
	User   UserType = "user"
	Player UserType = "player"
)

type UserAuth struct {
	ID       string    `json:"id"`
	Role     role.Role `json:"role"`
	UserType UserType  `json:"user_type"`
	ClientID string    `json:"client_id"`
}

type RegisterUserCommand struct {
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Phone      *string `json:"phone"`
}

type LoginUserCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ClientIP string `json:"-"`
}

type RefreshTokenCommand struct {
	RefreshToken string `json:"refresh_token"`
	ClientIP     string `json:"-"`
}

type LoginUserResult struct {
	FirstName        string `json:"first_name"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresAt        int64  `json:"expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}
