package transport

import (
	"github.com/daniildddd/maestro/internal/core/domain"
)

func usersResponseFromDomain(users []domain.User) []UserResponse {
	resp := make([]UserResponse, 0, len(users))

	for _, u := range users {
		resp = append(resp, userResponseFromDomain(u))
	}

	return resp
}

func userResponseFromDomain(u domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
