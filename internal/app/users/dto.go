package users

import "time"

type UserResponse struct {
	Id              uint      `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailVerifiedAt time.Time `json:"email_verified_at"`
	GroupId         uint      `json:"group_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}