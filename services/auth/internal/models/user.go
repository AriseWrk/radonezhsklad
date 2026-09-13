package models

import (
"time"

"github.com/google/uuid"
)

type User struct {
ID           uuid.UUID `json:"id"`
Email        string    `json:"email"`
PasswordHash string    `json:"-"`
FullName     string    `json:"full_name"`
LastName     string    `json:"last_name"`
FirstName    string    `json:"first_name"`
MiddleName   string    `json:"middle_name"`
Phone        *string   `json:"phone,omitempty"`
Login        *string   `json:"login,omitempty"`
Description  *string   `json:"description,omitempty"`
Role         string    `json:"role"`
IsActive     bool      `json:"is_active"`
CreatedAt    time.Time `json:"created_at"`
UpdatedAt    time.Time `json:"updated_at"`
}