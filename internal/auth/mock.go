package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type mockAuth struct{}

var mockClaims = jwt.MapClaims{
	"aud": "test-aud",
	"iss": "test-aud",
	"sub": int64(1),
	"exp": time.Now().Add(time.Hour).Unix(),
}

var secret = "test"

func NewMockAuth() *mockAuth {
	return &mockAuth{}
}

func (a *mockAuth) GenerateToken(claims jwt.Claims) (string, error) {
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, mockClaims)
	return jwt.SignedString([]byte(secret))
}

func (a *mockAuth) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
}
