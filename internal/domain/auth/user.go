package auth

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type User struct {
	id       uuid.UUID
	email    string
	username string
	role     string
}

func New(id uuid.UUID, email string, username string, role string) (User, error) {
	if id == uuid.Nil {
		return User{}, fmt.Errorf("invalid user ID")
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return User{}, fmt.Errorf("invalid user email")
	}

	username = strings.TrimSpace(username)
	if username == "" || len(username) > 32 || len(username) < 4 {
		return User{}, fmt.Errorf("invalid user username")
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return User{}, fmt.Errorf("invalid user role")
	}

	return User{
		id:       id,
		email:    email,
		username: username,
		role:     role,
	}, nil
}

func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Email() string {
	return u.email
}

func (u User) Username() string {
	return u.username
}

func (u User) Role() string {
	return u.role
}
