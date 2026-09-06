package auth

import (
	"fmt"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/lib/valid"
	"github.com/google/uuid"
)

type User struct {
	id    uuid.UUID
	email string
	role  string
}

func NewUser(id uuid.UUID, email string, role string) (User, error) {
	if id == uuid.Nil {
		return User{}, fmt.Errorf("id cannot be empty")
	}

	email, err := valid.NormalizeEmail(email)
	if err != nil {
		return User{}, fmt.Errorf("invalid email: %w", err)
	}

	role, err = valid.NormalizeValue(role)
	if err != nil {
		return User{}, fmt.Errorf("invalid role: %w", err)
	}

	if role != "user" && role != "admin" {
		return User{}, fmt.Errorf("invalid role %q: %w", role, errs.ErrInvalidInput)
	}

	return User{
		id:    id,
		email: email,
		role:  role,
	}, nil
}

func (u User) ID() uuid.UUID { return u.id }
func (u User) Email() string { return u.email }
func (u User) Role() string  { return u.role }
