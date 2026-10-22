package dto

import (
	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

func NewUserDTO(id uuid.UUID, email string, role string) User {
	return User{
		Id:    id,
		Email: email,
		Role:  role,
	}
}

func DomainToUserDTO(user auth.User) User {
	return User{
		Id:    user.ID(),
		Email: user.Email(),
		Role:  user.Role(),
	}
}
