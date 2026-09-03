package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateTokenFromUser(jwtSecret string, userID string) (string, error) {
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(8 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.New("couldnt generate token")
	}
	return ss, nil
}
