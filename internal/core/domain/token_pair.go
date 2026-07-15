package domain

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	Username     string
	ExpiresAt    time.Time
}

func NewTokenPair(
	accessToken string,
	refreshToken string,
	username string,
	expiresAt time.Time,
) TokenPair {
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     username,
		ExpiresAt:    expiresAt,
	}
}
