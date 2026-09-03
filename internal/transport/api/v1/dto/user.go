package dto

import (
	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Role  string    `json:"role"`
}

func NewUserDTO(id uuid.UUID, email string, name string, role string) User {
	return User{
		Id:    id,
		Email: email,
		Name:  name,
		Role:  role,
	}
}

func DomainToUserDTO(user auth.User) User {
	return User{
		Id:    user.ID(),
		Email: user.Email(),
		Name:  user.Username(),
		Role:  user.Role(),
	}
}
