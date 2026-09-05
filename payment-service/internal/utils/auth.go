package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrMissingUserID = errors.New("missing user id")
var ErrInvalidUserID = errors.New("invalid user id")

func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		return uuid.Nil, ErrMissingUserID
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrInvalidUserID
	}
	return parsedUserID, nil
}
