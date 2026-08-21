package auth

import (
	"github.com/imsab23/platform-be/pkg/util/error"
)

var (
	ErrInvalidCredentials = error.New("AUTH0000", "Invalid credentials.")
)

type UserAuth struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	UserType string `json:"user_type"`
	ClientID string `json:"client_id"`
}

type LoginUserCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResult struct {
	Token string `json:"token"`
}
