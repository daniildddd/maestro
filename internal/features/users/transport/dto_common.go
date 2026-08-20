package transport

import (
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type UserDTOResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func userDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func usersDTOFromDomains(users []domain.User) []UserDTOResponse {
	usersDTO := make([]UserDTOResponse, 0, len(users))

	for _, user := range users {
		usersDTO = append(usersDTO, userDTOFromDomain(user))
	}

	return usersDTO
}
