package api_auth

import (
	"github.com/miniyus/keyword-search-backend/internal/utils"
)

type SignUp struct {
	Username        string `json:"username" validate:"required"`
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
}

type SignIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type TokenInfo struct {
	Token     string         `json:"token"`
	ExpiresAt utils.JsonTime `json:"expires_at"`
	ExpiresIn int            `json:"expires_in"`
}

type SignUpResponse struct {
	UserId uint `json:"user_id"`
	TokenInfo
}

type ResetPasswordStruct struct {
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}
