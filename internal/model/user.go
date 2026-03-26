package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleGuest     Role = "guest"
	RoleUser      Role = "user"
	RoleOrganizer Role = "organizer"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type Status string

const (
	StatusActive              Status = "active"
	StatusBlocked             Status = "blocked"
	StatusPendingVerification Status = "pending_verification"
)

type User struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Email         string    `db:"email" json:"email"`
	PasswordHash  string    `db:"password_hash" json:"-"`
	Name          string    `db:"name" json:"name"`
	Role          Role      `db:"role" json:"role"`
	Status        Status    `db:"status" json:"status"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ValidateEmailRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}
