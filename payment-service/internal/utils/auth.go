package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrMissingMerchantID = errors.New("missing user id")
var ErrInvalidMerchantID = errors.New("invalid user id")

func GetMerchantID(c *gin.Context) (uuid.UUID, error) {
	userID := c.GetHeader("X-Merchant-ID")
	if userID == "" {
		return uuid.Nil, ErrMissingMerchantID
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrInvalidMerchantID
	}
	return parsedUserID, nil
}
