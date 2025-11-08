package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenicator struct {
	secret string
	aud    string
	iss    string
}

func NewJWTAuthenicator(secret, aud, iss string) *JWTAuthenicator {
	return &JWTAuthenicator{
		secret, iss, aud,
	}
}

func (a *JWTAuthenicator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *JWTAuthenicator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(a.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods(([]string{jwt.SigningMethodES256.Name})),
	)
}
